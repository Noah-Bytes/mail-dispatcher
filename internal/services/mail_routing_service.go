package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"mail-dispatcher/internal/models"

	"gorm.io/gorm"
)

// MailRoutingService 邮件路由服务
type MailRoutingService struct {
	db            *gorm.DB
	senderService *SenderService
	logService    *LogService
}

// NewMailRoutingService 创建邮件路由服务
func NewMailRoutingService(db *gorm.DB, senderService *SenderService, logService *LogService) *MailRoutingService {
	return &MailRoutingService{
		db:            db,
		senderService: senderService,
		logService:    logService,
	}
}

// UpdateSenderService 更新发送服务引用
func (s *MailRoutingService) UpdateSenderService(senderService *SenderService) {
	s.senderService = senderService
}

// ProcessEmail 处理新邮件
func (s *MailRoutingService) ProcessEmail(email models.Email, accountID uint) error {
	// 检查是否已处理过
	var existingLog models.MailLog
	if err := s.db.Where("message_id = ? AND account_id = ?", email.MessageID, accountID).First(&existingLog).Error; err == nil {
		log.Printf("邮件已处理过，跳过: %s", email.MessageID)
		return nil
	}

	// 解析邮件主题
	_, targetName, err := s.parseSubject(email.Subject)
	if err != nil {
		log.Printf("解析邮件主题失败: %v (主题: '%s')", err, email.Subject)
		return s.logFailedEmail(email, accountID, "解析主题失败: "+err.Error())
	}

	// 查找转发目标
	var target models.ForwardTarget
	if err := s.db.Where("name = ?", targetName).First(&target).Error; err != nil {
		log.Printf("未找到匹配的转发目标: %s", targetName)
		return s.logFailedEmail(email, accountID, "未找到匹配的转发目标: "+targetName)
	}

	// 使用发送服务发送邮件
	if err := s.senderService.SendEmail(email, target.Email, accountID); err != nil {
		log.Printf("转发邮件失败: %v", err)
		return s.logFailedEmail(email, accountID, "转发失败: "+err.Error())
	}

	// 记录成功日志
	return s.logSuccessfulEmail(email, accountID, target.Email)
}

// parseSubject 解析邮件主题
func (s *MailRoutingService) parseSubject(subject string) (keyword, targetName string, err error) {
	// 按照 "关键字 - 转发对象名称" 格式解析
	parts := strings.SplitN(subject, " - ", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("邮件主题格式不正确，应为 '关键字 - 转发对象名称'")
	}

	keyword = strings.TrimSpace(parts[0])
	targetName = strings.TrimSpace(parts[1])

	if keyword == "" || targetName == "" {
		return "", "", fmt.Errorf("关键字或转发对象名称为空")
	}

	return keyword, targetName, nil
}

// logSuccessfulEmail 记录成功转发的邮件
func (s *MailRoutingService) logSuccessfulEmail(email models.Email, accountID uint, forwardTo string) error {
	now := time.Now()
	log := models.MailLog{
		AccountID:   accountID,
		MessageID:   email.MessageID,
		Subject:     email.Subject,
		From:        email.From,
		To:          email.To,
		ReceivedAt:  email.ReceivedAt,
		ForwardTo:   forwardTo,
		Status:      "forwarded",
		ForwardedAt: &now,
	}

	return s.db.Create(&log).Error
}

// logFailedEmail 记录失败的邮件
func (s *MailRoutingService) logFailedEmail(email models.Email, accountID uint, errorMsg string) error {
	log := models.MailLog{
		AccountID:  accountID,
		MessageID:  email.MessageID,
		Subject:    email.Subject,
		From:       email.From,
		To:         email.To,
		ReceivedAt: email.ReceivedAt,
		Status:     "failed",
		Error:      errorMsg,
	}

	return s.db.Create(&log).Error
}
