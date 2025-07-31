package services

import (
	"fmt"
	"log"

	"mail-dispatcher/internal/config"
	"mail-dispatcher/internal/mail"
	"mail-dispatcher/internal/models"

	"gorm.io/gorm"
)

// SenderService 发送服务
type SenderService struct {
	db     *gorm.DB
	config *config.Config
}

// NewSenderService 创建发送服务
func NewSenderService(db *gorm.DB, cfg *config.Config) *SenderService {
	return &SenderService{
		db:     db,
		config: cfg,
	}
}

// SendEmail 发送邮件
func (s *SenderService) SendEmail(email models.Email, toEmail string, accountID uint) error {
	// 动态获取账户信息
	var account models.MailAccount
	if err := s.db.First(&account, accountID).Error; err != nil {
		return fmt.Errorf("未找到账户: %d", accountID)
	}

	// 动态创建邮件客户端
	mailClient, err := s.createMailClient(account)
	if err != nil {
		return fmt.Errorf("创建邮件客户端失败: %v", err)
	}

	// 确保客户端被正确关闭
	defer func() {
		if err := mailClient.Stop(); err != nil {
			log.Printf("停止邮件客户端失败 (账户ID: %d): %v", accountID, err)
		}
	}()

	// 使用邮件客户端发送邮件
	if err := mailClient.SendEmail(email, toEmail); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	log.Printf("邮件发送成功: %s -> %s (账户ID: %d)", email.Subject, toEmail, accountID)
	return nil
}

// SendRawEmail 发送原始邮件数据
func (s *SenderService) SendRawEmail(rawData []byte, toEmail string, accountID uint) error {
	// 动态获取账户信息
	var account models.MailAccount
	if err := s.db.First(&account, accountID).Error; err != nil {
		return fmt.Errorf("未找到账户: %d", accountID)
	}

	// 动态创建邮件客户端
	mailClient, err := s.createMailClient(account)
	if err != nil {
		return fmt.Errorf("创建邮件客户端失败: %v", err)
	}

	// 确保客户端被正确关闭
	defer func() {
		if err := mailClient.Stop(); err != nil {
			log.Printf("停止邮件客户端失败 (账户ID: %d): %v", accountID, err)
		}
	}()

	// 使用邮件客户端发送原始邮件
	if err := mailClient.SendRawEmail(rawData, toEmail); err != nil {
		return fmt.Errorf("发送原始邮件失败: %v", err)
	}

	log.Printf("原始邮件发送成功: %s (账户ID: %d)", toEmail, accountID)
	return nil
}

// createMailClient 动态创建邮件客户端
func (s *SenderService) createMailClient(account models.MailAccount) (*mail.MailClient, error) {
	mailClient := mail.NewMailClient(s.config)

	// 初始化邮件客户端
	config := mail.Config{
		AccountID: account.ID,
		Provider:  "mail", // 统一使用 mail 类型
		Address:   account.Address,
		Username:  account.Username,
		Password:  account.Password,
		Server:    account.Server,
		Settings:  account.Settings,
		LastUID:   account.LastUID,
	}

	if err := mailClient.Init(config); err != nil {
		return nil, err
	}

	return mailClient, nil
}
