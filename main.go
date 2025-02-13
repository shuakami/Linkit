package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"linkit/internal/delivery/http"
	"linkit/internal/domain"
	"linkit/internal/infrastructure/cache"
	"linkit/internal/infrastructure/database"
	"linkit/internal/infrastructure/logger"
	"linkit/internal/repository"
	"linkit/internal/usecase"
	"linkit/pkg/utils"
)

func init() {
	// 加载配置文件
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	// 初始化IP搜索器
	if err := utils.InitIPSearcher("ip2region/ip2region.xdb"); err != nil {
		log.Printf("Warning: Failed to initialize IP searcher: %v", err)
	}
}

func main() {
	// 初始化日志
	zapLogger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer zapLogger.Sync()

	sugar := logger.NewSugaredLogger(zapLogger)
	defer sugar.Sync()

	// 初始化数据库连接
	db, err := database.NewPostgresDB()
	if err != nil {
		sugar.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库结构
	if err := db.AutoMigrate(&domain.ShortLink{}, &domain.RedirectRule{}, &domain.ClickLog{}); err != nil {
		sugar.Fatalf("Failed to migrate database: %v", err)
	}
	sugar.Info("Database migrated successfully")

	// 初始化Redis连接
	redisClient, err := cache.NewRedisClient()
	if err != nil {
		sugar.Fatalf("Failed to connect to redis: %v", err)
	}

	// 清理所有缓存
	if err := redisClient.FlushDB(context.Background()).Err(); err != nil {
		sugar.Fatalf("Failed to flush redis: %v", err)
	}
	sugar.Info("Cleared all cache")

	// 初始化仓储层
	shortLinkRepo := repository.NewShortLinkRepository(db, redisClient)

	// 初始化用例层
	shortLinkUseCase := usecase.NewShortLinkUseCase(shortLinkRepo)

	// 初始化处理器
	shortLinkHandler := http.NewShortLinkHandler(shortLinkUseCase)

	// 设置gin模式
	gin.SetMode(viper.GetString("server.mode"))

	// 创建gin实例
	r := gin.Default()

	// 注册路由
	http.RegisterRoutes(r, shortLinkHandler)

	// 启动服务器
	addr := fmt.Sprintf(":%d", viper.GetInt("server.port"))
	if err := r.Run(addr); err != nil {
		sugar.Fatalf("Failed to start server: %v", err)
	}

	// 关闭IP搜索器
	defer utils.CloseIPSearcher()
}
