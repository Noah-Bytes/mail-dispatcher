package services

import (
	"fmt"
	"log"
	"time"

	"mail-dispatcher/internal/config"
	"mail-dispatcher/internal/mail"
	"mail-dispatcher/internal/models"

	"gorm.io/gorm"
)

// SchedulerService 调度器服务
type SchedulerService struct {
	db                 *gorm.DB
	mailRoutingService *MailRoutingService
	config             *config.Config
	stopChan           chan bool
}

// NewSchedulerService 创建调度器服务
func NewSchedulerService(db *gorm.DB, mailRoutingService *MailRoutingService, cfg *config.Config) *SchedulerService {
	return &SchedulerService{
		db:                 db,
		mailRoutingService: mailRoutingService,
		config:             cfg,
		stopChan:           make(chan bool),
	}
}

// Start 启动调度器
func (s *SchedulerService) Start() {
	// 先执行一次轮询
	s.pollAllAccounts()

	go s.pollingLoop()
	log.Println("调度器服务已启动")
}

// Stop 停止调度器
func (s *SchedulerService) Stop() {
	close(s.stopChan)
	log.Println("调度器服务已停止")
}

// pollingLoop 轮询循环
func (s *SchedulerService) pollingLoop() {
	// 使用配置中的轮询间隔
	fmt.Println(s.config.Mail)
	pollingInterval := time.Duration(s.config.Mail.PollingInterval) * time.Second
	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()

	log.Printf("调度器使用轮询间隔: %v", pollingInterval)

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.pollAllAccounts()
		}
	}
}

// pollAllAccounts 轮询所有账户
func (s *SchedulerService) pollAllAccounts() {
	// 每次轮询时动态获取活跃账户
	accounts, err := s.getActiveAccounts()
	if err != nil {
		log.Printf("获取活跃账户失败: %v", err)
		return
	}

	log.Printf("开始轮询，共有 %d 个活跃账户", len(accounts))

	for _, account := range accounts {
		log.Printf("轮询账户: %s (账户ID: %d)", account.Address, account.ID)
		go s.pollAccount(account)
	}
}

// pollAccount 轮询单个账户
func (s *SchedulerService) pollAccount(account models.MailAccount) {
	log.Printf("开始轮询账户: %s (账户ID: %d)", account.Address, account.ID)

	// 动态创建邮件客户端
	mailClient, err := s.createMailClient(account)
	if err != nil {
		log.Printf("创建邮件客户端失败 (账户ID: %d): %v", account.ID, err)
		return
	}

	// 确保客户端被正确关闭
	defer func() {
		if err := mailClient.Stop(); err != nil {
			log.Printf("停止邮件客户端失败 (账户ID: %d): %v", account.ID, err)
		}
	}()

	emails, err := mailClient.FetchNewEmails()
	if err != nil {
		log.Printf("轮询邮件客户端失败 (账户ID: %d): %v", account.ID, err)
		return
	}

	log.Printf("账户 %s 获取到 %d 封新邮件", account.Address, len(emails))

	// 处理每封邮件
	for _, email := range emails {
		if err := s.mailRoutingService.ProcessEmail(email, account.ID); err != nil {
			log.Printf("处理邮件失败: %v", err)
		}
	}
}

// getActiveAccounts 获取所有活跃账户
func (s *SchedulerService) getActiveAccounts() ([]models.MailAccount, error) {
	var accounts []models.MailAccount
	if err := s.db.Where("is_active = ?", true).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// createMailClient 根据账户动态创建邮件客户端
func (s *SchedulerService) createMailClient(account models.MailAccount) (*mail.MailClient, error) {
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
