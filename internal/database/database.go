package database

import (
	"fmt"
	"time"

	"enterprise-blog/internal/config"
	"enterprise-blog/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB 是全局的 GORM 数据库连接
var DB *gorm.DB

func Init() error {
	dsn := config.AppConfig.Database.DSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池（通过底层 *sql.DB）
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	l := logger.GetLogger()
	l.Info().Msg("Database connected successfully")

	return nil
}

func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

