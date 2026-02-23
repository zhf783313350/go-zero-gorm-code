package svc

import (
	"accesscontrol/internal/config"
	"accesscontrol/internal/repository"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/syncx"
)

type ServiceContext struct {
	Config      config.Config
	DB          sqlx.SqlConn
	UserRepo    repository.UserRepository
	Redis       *redis.Redis
	RateLimiter *limit.TokenLimiter
	SingleGroup syncx.SingleFlight
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 构建 PostgreSQL 连接字符串
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)

	// 连接数据库 (使用 go-zero/sqlx)
	db := sqlx.NewSqlConn("postgres", dsn)

	rawDB, err := db.RawDB()
	if err == nil {
		rawDB.SetMaxOpenConns(c.Database.MaxOpenConns)
		rawDB.SetMaxIdleConns(c.Database.MaxIdleConns)
		rawDB.SetConnMaxLifetime(time.Duration(c.Database.ConnMaxLifetime) * time.Second)
	}

	// 运行数据库迁移 (golang-migrate 需要 URL 格式的 DSN)
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
		// go-zero Redis底层暂无对外暴露 PoolSize / MinIdleConns 直接映射到 New 初始化，但如果引入 go-redis 可设置。
		// 这里暂且保持现有接口调用，如果您以后换成直接用 go-redis 的客户端可以如下传递。
	})

	// 初始化 RateLimiter (使用配置项，解除硬编码死锁)
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
