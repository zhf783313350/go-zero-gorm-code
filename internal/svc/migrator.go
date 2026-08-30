package svc

import (
	"embed"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/zeromicro/go-zero/core/logx"
)

// 🔥 核心修复：添加 go:embed 指令，明确告诉 Go 编译器在编译时将 migrations 目录下的所有 sql 文件打包进二进制
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(dsn string) error {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}

	// 自动更新到最新版本
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	logx.Info("Database migrations completed successfully")
	return nil
}
