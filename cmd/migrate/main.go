package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"enterprise-blog/internal/config"
	"enterprise-blog/internal/database"
	"enterprise-blog/pkg/logger"
)

type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

func main() {
	// 加载配置
	if err := config.Load(); err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 初始化日志
	if err := logger.Init("info", ""); err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}

	// 初始化数据库（使用 GORM）
	if err := database.Init(); err != nil {
		l := logger.GetLogger()
		l.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close()

	// 创建迁移表
	if err := createMigrationsTable(); err != nil {
		l := logger.GetLogger()
		l.Fatal().Err(err).Msg("Failed to create migrations table")
	}

	// 加载迁移文件
	migrations, err := loadMigrations()
	if err != nil {
		l := logger.GetLogger()
		l.Fatal().Err(err).Msg("Failed to load migrations")
	}

	// 获取已执行的迁移
	executed, err := getExecutedMigrations()
	if err != nil {
		l := logger.GetLogger()
		l.Fatal().Err(err).Msg("Failed to get executed migrations")
	}

	// 执行未执行的迁移
	for _, migration := range migrations {
		if executed[migration.Version] {
			l := logger.GetLogger()
			l.Info().Int("version", migration.Version).Msg("Migration already executed, skipping")
			continue
		}

		l := logger.GetLogger()
		l.Info().Int("version", migration.Version).Str("name", migration.Name).Msg("Running migration")

		tx := database.DB.Begin()
		if tx.Error != nil {
			l := logger.GetLogger()
			l.Fatal().Err(tx.Error).Msg("Failed to begin transaction")
		}

		// 执行 up 迁移
		if err := tx.Exec(migration.Up).Error; err != nil {
			tx.Rollback()
			l := logger.GetLogger()
			l.Fatal().Err(err).Int("version", migration.Version).Msg("Failed to execute migration")
		}

		// 记录迁移
		if err := tx.Exec("INSERT INTO schema_migrations (version, name) VALUES ($1, $2)", migration.Version, migration.Name).Error; err != nil {
			tx.Rollback()
			l := logger.GetLogger()
			l.Fatal().Err(err).Msg("Failed to record migration")
		}

		if err := tx.Commit().Error; err != nil {
			l := logger.GetLogger()
			l.Fatal().Err(err).Msg("Failed to commit transaction")
		}

		l2 := logger.GetLogger()
		l2.Info().Int("version", migration.Version).Msg("Migration completed")
	}

	l := logger.GetLogger()
	l.Info().Msg("All migrations completed")
}

func createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	return database.DB.Exec(query).Error
}

func loadMigrations() ([]Migration, error) {
	// 直接从项目根目录下的 migrations 目录读取 SQL 文件
	migrationsDir := filepath.Join("migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	migrationsMap := make(map[int]*Migration)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		parts := strings.Split(entry.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		var version int
		var direction string
		var name string

		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			continue
		}

		filename := entry.Name()
		if strings.HasSuffix(filename, ".up.sql") {
			direction = "up"
			name = strings.TrimSuffix(strings.TrimPrefix(filename, fmt.Sprintf("%03d_", version)), ".up.sql")
		} else if strings.HasSuffix(filename, ".down.sql") {
			direction = "down"
			name = strings.TrimSuffix(strings.TrimPrefix(filename, fmt.Sprintf("%03d_", version)), ".down.sql")
		} else {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return nil, err
		}

		if migrationsMap[version] == nil {
			migrationsMap[version] = &Migration{
				Version: version,
				Name:    name,
			}
		}

		if direction == "up" {
			migrationsMap[version].Up = string(content)
		} else {
			migrationsMap[version].Down = string(content)
		}
	}

	var migrations []Migration
	for _, m := range migrationsMap {
		migrations = append(migrations, *m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func getExecutedMigrations() (map[int]bool, error) {
	var versions []int
	if err := database.DB.Raw("SELECT version FROM schema_migrations").Scan(&versions).Error; err != nil {
		return nil, err
	}

	executed := make(map[int]bool)
	for _, v := range versions {
		executed[v] = true
	}
	return executed, nil
}
