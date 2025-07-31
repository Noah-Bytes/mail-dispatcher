package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"mail-dispatcher/internal/models"
	"mail-dispatcher/internal/services"
)

// LogController 邮件日志控制器
type LogController struct {
	logService *services.LogService
}

// NewLogController 创建邮件日志控制器
func NewLogController(logService *services.LogService) *LogController {
	return &LogController{logService: logService}
}

// GetLogs 获取邮件日志
func (c *LogController) GetLogs(ctx *gin.Context) {
	// 获取查询参数
	status := ctx.Query("status")
	accountIDStr := ctx.Query("account_id")
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的limit参数"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的offset参数"})
		return
	}

	var logs []models.MailLog
	var logsErr error

	if accountIDStr != "" {
		accountID, parseErr := strconv.ParseUint(accountIDStr, 10, 32)
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的account_id参数"})
			return
		}
		logs, logsErr = c.logService.GetLogsByAccount(uint(accountID), limit, offset)
	} else if status != "" {
		logs, logsErr = c.logService.GetLogsByStatus(status, limit, offset)
	} else {
		// 获取所有日志（这里需要添加一个方法）
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请指定status或account_id参数"})
		return
	}

	if logsErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取日志失败: " + logsErr.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": len(logs),
	})
}

// GetFailedLogs 获取失败的日志
func (c *LogController) GetFailedLogs(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的limit参数"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的offset参数"})
		return
	}

	logs, err := c.logService.GetFailedLogs(limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败日志失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": len(logs),
	})
}

// GetSuccessfulLogs 获取成功的日志
func (c *LogController) GetSuccessfulLogs(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的limit参数"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的offset参数"})
		return
	}

	logs, err := c.logService.GetSuccessfulLogs(limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取成功日志失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": len(logs),
	})
}

// GetLogsByDateRange 根据日期范围获取日志
func (c *LogController) GetLogsByDateRange(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date和end_date参数不能为空"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的start_date格式，应为YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的end_date格式，应为YYYY-MM-DD"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的limit参数"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的offset参数"})
		return
	}

	logs, err := c.logService.GetLogsByDateRange(startDate, endDate, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取日志失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": len(logs),
	})
}

// GetLogsStats 获取日志统计信息
func (c *LogController) GetLogsStats(ctx *gin.Context) {
	total, err := c.logService.GetLogsCount()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取日志总数失败: " + err.Error()})
		return
	}

	failedCount, err := c.logService.GetLogsCountByStatus("failed")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败日志数量失败: " + err.Error()})
		return
	}

	successCount, err := c.logService.GetLogsCountByStatus("forwarded")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取成功日志数量失败: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"total":      total,
		"failed":     failedCount,
		"successful": successCount,
	})
}
