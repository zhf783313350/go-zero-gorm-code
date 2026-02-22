package svc

import (
	"accesscontrol/internal/config"
	"accesscontrol/internal/repository"
	"fmt"

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
	})

	// 初始化 RateLimiter (100 req/s)
	limiter := limit.NewTokenLimiter(100, 100, rds, "api-rate-limit")

	return &ServiceContext{
		Config:      c,
		DB:          db,
		UserRepo:    repository.NewUserRepository(db),
		Redis:       rds,
		RateLimiter: limiter,
		SingleGroup: syncx.NewSingleFlight(),
	}
}
