package routes

import (
	"mail-dispatcher/internal/controllers"
	"mail-dispatcher/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes 设置路由
func SetupRoutes(router *gin.Engine, db *gorm.DB, logService *services.LogService) {
	// 创建控制器
	targetController := controllers.NewTargetController(db)
	accountController := controllers.NewAccountController(db)
	logController := controllers.NewLogController(logService)

	// API路由组
	api := router.Group("/api/v1")
	{
		// 转发目标管理
		targets := api.Group("/targets")
		{
			targets.GET("", targetController.GetTargets)
			targets.GET("/:id", targetController.GetTarget)
			targets.POST("", targetController.CreateTarget)
			targets.PUT("/:id", targetController.UpdateTarget)
			targets.DELETE("/:id", targetController.DeleteTarget)
		}

		// 邮箱账户管理
		accounts := api.Group("/accounts")
		{
			accounts.GET("", accountController.GetAccounts)
			accounts.GET("/:id", accountController.GetAccount)
			accounts.POST("", accountController.CreateAccount)
			accounts.PUT("/:id", accountController.UpdateAccount)
			accounts.DELETE("/:id", accountController.DeleteAccount)
			accounts.PUT("/:id/toggle", accountController.ToggleAccountStatus)
		}

		// 邮件日志管理
		logs := api.Group("/logs")
		{
			logs.GET("", logController.GetLogs)
			logs.GET("/failed", logController.GetFailedLogs)
			logs.GET("/successful", logController.GetSuccessfulLogs)
			logs.GET("/range", logController.GetLogsByDateRange)
			logs.GET("/stats", logController.GetLogsStats)
		}
	}

	// 健康检查
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "healthy",
		})
	})

	// 根路径
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "邮件转发系统API",
			"version": "1.0.0",
			"endpoints": gin.H{
				"targets":  "/api/v1/targets",
				"accounts": "/api/v1/accounts",
				"logs":     "/api/v1/logs",
				"health":   "/ping",
			},
		})
	})
}
