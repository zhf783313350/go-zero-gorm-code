package svc

import (
	"accesscontrol/internal/config"
	"accesscontrol/internal/repository"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config      config.Config
	DB          *gorm.DB
	UserRepo    repository.UserRepository
	Redis       *redis.Redis
	RateLimiter *limit.TokenLimiter
	SingleGroup syncx.SingleFlight
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 构建 PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)

	// 连接数据库 (使用 GORM)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get underlying sql.DB: %v", err))
	}
	sqlDB.SetMaxOpenConns(c.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(c.Database.ConnMaxLifetime) * time.Second)

	// 运行数据库迁移 (golang-migrate 使用 URL 格式 DSN)
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

	// 初始化 Redis
	rds := redis.New(c.Redis.Host, func(r *redis.Redis) {
		r.Type = redis.NodeType
		r.Pass = c.Redis.Password
	})

	// 初始化 RateLimiter
	limiter := limit.NewTokenLimiter(c.RateLimiter.Rate, c.RateLimiter.Burst, rds, "api-rate-limit")

	return &ServiceContext{
		Config:      c,
		DB:          db,
		UserRepo:    repository.NewUserRepository(db),
		Redis:       rds,
		RateLimiter: limiter,
		SingleGroup: syncx.NewSingleFlight(),
	}
}
