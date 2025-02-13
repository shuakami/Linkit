package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../configs")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Command required: up or down")
	}

	command := os.Args[1]

	// 构建数据库连接字符串
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		viper.GetString("database.user"),
		viper.GetString("database.password"),
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.dbname"),
		viper.GetString("database.sslmode"),
	)

	// 创建迁移实例
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	// 执行迁移
	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("Successfully migrated up")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println("Successfully migrated down")
	default:
		log.Fatal("Invalid command. Use 'up' or 'down'")
	}
}
