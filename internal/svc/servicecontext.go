package svc

import (
	"accesscontrol/internal/config"
	"accesscontrol/internal/domain"
	"accesscontrol/internal/event"
	"accesscontrol/internal/repository"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	casbin "github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

type ServiceContext struct {
	Config      config.Config
	DB          *gorm.DB
	UserRepo    domain.UserRepository
	Redis       *redis.Redis
	RateLimiter *limit.TokenLimiter
	SingleGroup syncx.SingleFlight
	EventBus    *event.EventBus
	Enforcer    *casbin.Enforcer
}

func NewServiceContext(c config.Config) *ServiceContext {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		AllowGlobalUpdate: false,
		Plugins:           map[string]gorm.Plugin{},
		Logger:            logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get underlying sql.DB: %v", err))
	}
	sqlDB.SetMaxOpenConns(c.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(c.Database.ConnMaxLifetime) * time.Second)

	migrationDsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
		c.Database.SSLMode,
	)
	if err := RunMigrations(migrationDsn); err != nil {
		panic(fmt.Sprintf("failed to run database migrations: %v", err))
	}

	rds := redis.New(c.Redis.Host, func(r *redis.Redis) {
		r.Type = redis.NodeType
		r.Pass = c.Redis.Password
	})

	limiter := limit.NewTokenLimiter(c.RateLimiter.Rate, c.RateLimiter.Burst, rds, "api-rate-limit")

	// 初始化异步事件总线（缓冲 1000 个事件，防止突发流量堆积）
	bus := event.NewEventBus(1000)

	// 初始化 Casbin RBAC 权限管理（采用内存 Model + DB Adapter，零部署成本）
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
	cbModel, err := casbinmodel.NewModelFromString(modelText)
	if err != nil {
		panic(fmt.Sprintf("failed to parse casbin model: %v", err))
	}

	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize casbin gorm adapter: %v", err))
	}

	enforcer, err := casbin.NewEnforcer(cbModel, adapter)
	if err != nil {
		panic(fmt.Sprintf("failed to create casbin enforcer: %v", err))
	}

	if err := enforcer.LoadPolicy(); err != nil {
		panic(fmt.Sprintf("failed to load casbin policies: %v", err))
	}

	// 注入默认权限规则 (Seed 数据)
	// 允许 admin 角色访问所有的受保护接口
	if len(enforcer.GetPolicy()) == 0 {
		_, _ = enforcer.AddPolicy("admin", "/api/user/add", "POST")
		_, _ = enforcer.AddPolicy("admin", "/api/user/update", "POST")
		_, _ = enforcer.AddPolicy("admin", "/api/user/delete", "POST")
		_, _ = enforcer.AddPolicy("admin", "/api/user/list", "POST")
		// 允许普通用户角色访问所有接口
		_, _ = enforcer.AddPolicy("user", "/api/user/add", "POST")
		_, _ = enforcer.AddPolicy("user", "/api/user/update", "POST")
		_, _ = enforcer.AddPolicy("user", "/api/user/delete", "POST")
		_, _ = enforcer.AddPolicy("user", "/api/user/list", "POST")
		// 绑定用户 ID 1 到 admin 角色进行演示测试
		_, _ = enforcer.AddGroupingPolicy("1", "admin")
		_ = enforcer.SavePolicy()
	}

	// 自动将所有用户绑定到 user 角色（确保所有登录用户都能访问 list 接口）
	var userIDs []int64
	db.Raw("SELECT id FROM users").Scan(&userIDs)
	for _, userID := range userIDs {
		userIDStr := fmt.Sprintf("%d", userID)
		// 检查是否已绑定角色
		hasRole := false
		groupingPolicies := enforcer.GetGroupingPolicy()
		for _, policy := range groupingPolicies {
			if policy[0] == userIDStr {
				hasRole = true
				break
			}
		}
		if !hasRole {
			_, _ = enforcer.AddGroupingPolicy(userIDStr, "user")
		}
	}
	_ = enforcer.SavePolicy()

	return &ServiceContext{
		Config:      c,
		DB:          db,
		UserRepo:    repository.NewUserRepository(db),
		Redis:       rds,
		RateLimiter: limiter,
		SingleGroup: syncx.NewSingleFlight(),
		EventBus:    bus,
		Enforcer:    enforcer,
	}
}