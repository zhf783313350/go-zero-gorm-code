package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	// 数据库配置 (PostgreSQL)
	Database struct {
		Host            string
		Port            int
		User            string
		Password        string
		DBName          string
		SSLMode         string
		MaxOpenConns    int `json:",default=100"`
		MaxIdleConns    int `json:",default=10"`
		ConnMaxLifetime int `json:",default=3600"`
	}

	// Redis配置
	Redis struct {
		Host         string
		Password     string
		DB           int
		PoolSize     int `json:",default=100"`
		MinIdleConns int `json:",default=10"`
	}

	// JWT配置
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}

	// 全局限流配置
	RateLimiter struct {
		Rate  int `json:",default=2000"`
		Burst int `json:",default=3000"`
	}
}
