package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mail-dispatcher/internal/config"
	"mail-dispatcher/internal/models"
	"mail-dispatcher/internal/routes"
	"mail-dispatcher/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// load config
	cfg := config.LoadConfig()

	fmt.Println(cfg)

	// init database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("init database failed: %v", err)
	}

	// auto migrate database tables
	if err := db.AutoMigrate(&models.ForwardTarget{}, &models.MailAccount{}, &models.MailLog{}); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	// init services
	logService := services.NewLogService(db)

	// 初始化发送服务
	senderService := services.NewSenderService(db, cfg)

	// 初始化邮件路由服务
	mailRoutingService := services.NewMailRoutingService(db, senderService, logService)

	// 初始化调度器服务
	schedulerService := services.NewSchedulerService(db, mailRoutingService, cfg)

	// 启动调度器
	schedulerService.Start()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由
	router := gin.Default()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 设置路由
	routes.SetupRoutes(router, db, logService)

	// 启动HTTP服务器
	go func() {
		addr := cfg.Server.Host + ":" + cfg.Server.Port
		log.Printf("服务器启动在: %s", addr)
		if err := router.Run(addr); err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	// 停止调度器
	schedulerService.Stop()

	log.Println("服务器已关闭")
}

// initDatabase init database connection
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// get underlying sql.DB object
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// set connection pool parameters
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}
