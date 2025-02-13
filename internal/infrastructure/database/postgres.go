package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB 创建PostgreSQL数据库连接
func NewPostgresDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.user"),
		viper.GetString("database.password"),
		viper.GetString("database.dbname"),
		viper.GetString("database.sslmode"),
	)

	// 配置GORM日志
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // 慢SQL阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: false,       // 记录ErrRecordNotFound错误
			Colorful:                  true,        // 彩色打印
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
		DryRun: false, // 设置为true可以生成SQL而不执行
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置最大连接数
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	return db, nil
}
