package services

import (
	"time"

	"gorm.io/gorm"
	"mail-dispatcher/internal/models"
)

// LogService 日志服务
type LogService struct {
	db *gorm.DB
}

// NewLogService 创建日志服务
func NewLogService(db *gorm.DB) *LogService {
	return &LogService{db: db}
}

// CreateLog 创建日志记录
func (s *LogService) CreateLog(log *models.MailLog) error {
	return s.db.Create(log).Error
}

// GetLogsByAccount 根据账户获取日志
func (s *LogService) GetLogsByAccount(accountID uint, limit, offset int) ([]models.MailLog, error) {
	var logs []models.MailLog
	err := s.db.Where("account_id = ?", accountID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// GetLogsByStatus 根据状态获取日志
func (s *LogService) GetLogsByStatus(status string, limit, offset int) ([]models.MailLog, error) {
	var logs []models.MailLog
	err := s.db.Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// GetLogsByDateRange 根据日期范围获取日志
func (s *LogService) GetLogsByDateRange(startDate, endDate time.Time, limit, offset int) ([]models.MailLog, error) {
	var logs []models.MailLog
	err := s.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// GetFailedLogs 获取失败的日志
func (s *LogService) GetFailedLogs(limit, offset int) ([]models.MailLog, error) {
	return s.GetLogsByStatus("failed", limit, offset)
}

// GetSuccessfulLogs 获取成功的日志
func (s *LogService) GetSuccessfulLogs(limit, offset int) ([]models.MailLog, error) {
	return s.GetLogsByStatus("forwarded", limit, offset)
}

// GetLogsCount 获取日志总数
func (s *LogService) GetLogsCount() (int64, error) {
	var count int64
	err := s.db.Model(&models.MailLog{}).Count(&count).Error
	return count, err
}

// GetLogsCountByStatus 根据状态获取日志数量
func (s *LogService) GetLogsCountByStatus(status string) (int64, error) {
	var count int64
	err := s.db.Model(&models.MailLog{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CleanOldLogs 清理旧日志
func (s *LogService) CleanOldLogs(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return s.db.Where("created_at < ?", cutoffDate).Delete(&models.MailLog{}).Error
}
