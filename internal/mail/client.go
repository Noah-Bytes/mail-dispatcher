package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"

	"mail-dispatcher/internal/config"
	"mail-dispatcher/internal/models"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
)

// MailClient 邮件客户端（支持IMAP获取和SMTP发送）
type MailClient struct {
	config    Config
	client    *client.Client
	stopChan  chan bool
	appConfig *config.Config
}

// NewMailClient 创建新的邮件客户端
func NewMailClient(appConfig *config.Config) *MailClient {
	return &MailClient{
		stopChan:  make(chan bool),
		appConfig: appConfig,
	}
}

// Init 初始化邮件连接
func (c *MailClient) Init(config Config) error {
	c.config = config

	// 解析服务器地址
	server := config.Server
	if !strings.Contains(server, ":") {
		server += ":993" // 默认IMAPS端口
	}

	// 连接IMAP服务器
	imapClient, err := client.DialTLS(server, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return fmt.Errorf("连接IMAP服务器失败: %v", err)
	}

	// 登录
	if err := imapClient.Login(config.Username, config.Password); err != nil {
		imapClient.Logout()
		return fmt.Errorf("IMAP登录失败: %v", err)
	}

	c.client = imapClient
	log.Printf("邮件客户端初始化成功: %s", config.Address)
	return nil
}

// FetchNewEmails 获取新邮件
func (c *MailClient) FetchNewEmails() ([]models.Email, error) {
	// 检查连接状态，如果断开则重连
	if err := c.ensureConnection(); err != nil {
		return nil, fmt.Errorf("确保连接失败: %v", err)
	}

	// 选择收件箱
	_, err := c.client.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("选择收件箱失败: %v", err)
	}

	// 搜索未读邮件
	criteria := imap.NewSearchCriteria()
	criteria.Since = time.Now().AddDate(0, 0, -7) // 最近7天的邮件
	criteria.WithoutFlags = []string{imap.SeenFlag}

	uids, err := c.client.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("搜索邮件失败: %v", err)
	}

	if len(uids) == 0 {
		return []models.Email{}, nil
	}

	// 获取邮件内容 - 使用同步方式
	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchBody, imap.FetchUid}
	messages := make(chan *imap.Message, len(uids))

	// 同步获取邮件
	if err := c.client.Fetch(seqset, items, messages); err != nil {
		return nil, fmt.Errorf("获取邮件内容失败: %v", err)
	}

	// 收集邮件数据
	var emails []models.Email
	for msg := range messages {
		email, err := c.parseMessage(msg)
		if err != nil {
			log.Printf("解析邮件失败: %v", err)
			continue
		}
		emails = append(emails, email)
	}

	return emails, nil
}

// ensureConnection 确保连接可用，如果断开则重连
func (c *MailClient) ensureConnection() error {
	if c.client == nil {
		return c.reconnect()
	}

	// 尝试发送一个简单的命令来测试连接
	if err := c.client.Noop(); err != nil {
		return c.reconnect()
	}

	return nil
}

// reconnect 重新连接
func (c *MailClient) reconnect() error {
	// 关闭旧连接
	if c.client != nil {
		c.client.Logout()
		c.client = nil
	}

	// 解析服务器地址
	server := c.config.Server
	if !strings.Contains(server, ":") {
		server += ":993" // 默认IMAPS端口
	}

	// 使用配置中的重试参数
	maxRetryCount := 3
	retryInterval := 1
	if c.appConfig != nil {
		maxRetryCount = c.appConfig.Mail.MaxRetryCount
		retryInterval = c.appConfig.Mail.RetryInterval
	}

	// 连接IMAP服务器，添加重试机制
	var imapClient *client.Client
	var err error

	for i := 0; i < maxRetryCount; i++ {
		imapClient, err = client.DialTLS(server, &tls.Config{InsecureSkipVerify: true})
		if err == nil {
			break
		}
		time.Sleep(time.Duration(retryInterval) * time.Second)
	}

	if err != nil {
		return fmt.Errorf("连接IMAP服务器失败: %v", err)
	}

	// 登录，添加重试机制
	var loginErr error
	for i := 0; i < maxRetryCount; i++ {
		if err := imapClient.Login(c.config.Username, c.config.Password); err == nil {
			break
		} else {
			loginErr = err
		}
		time.Sleep(time.Duration(retryInterval) * time.Second)
	}

	if loginErr != nil {
		imapClient.Logout()
		return fmt.Errorf("IMAP登录失败: %v", loginErr)
	}

	c.client = imapClient
	return nil
}

// StartPushListener 启动推送监听（已禁用，只使用轮询）
func (c *MailClient) StartPushListener(callback func(models.Email)) error {
	// IDLE 推送已禁用，只使用轮询模式
	return fmt.Errorf("IDLE推送已禁用，使用轮询模式")
}

// Stop 停止Provider
func (c *MailClient) Stop() error {
	close(c.stopChan)
	if c.client != nil {
		return c.client.Logout()
	}
	return nil
}

// GetName 获取Provider名称
func (c *MailClient) GetName() string {
	return "IMAP"
}

// SendEmail 发送邮件
func (c *MailClient) SendEmail(email models.Email, toEmail string) error {
	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = c.config.Username
	headers["To"] = toEmail
	headers["Subject"] = email.Subject
	headers["Resent-From"] = c.config.Username
	headers["Resent-To"] = toEmail
	headers["X-Forwarded-By"] = "Mail-Dispatcher-System"

	// 如果有原始发件人信息，保留在邮件头中
	if email.From != "" {
		headers["Original-From"] = email.From
	}

	// 构建邮件内容
	var body bytes.Buffer

	// 写入邮件头
	for key, value := range headers {
		body.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	body.WriteString("\r\n")

	// 写入邮件正文
	if email.Body != "" {
		body.WriteString(email.Body)
	} else {
		body.WriteString("邮件内容")
	}

	// 解析SMTP服务器地址和端口
	smtpServer := c.getSMTPServer()
	smtpPort := c.getSMTPPort()

	// 尝试发送邮件，支持不同的连接方式
	err := c.sendMailWithFallback(smtpServer, smtpPort, toEmail, body.Bytes())
	if err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	log.Printf("IMAP Provider 邮件发送成功: %s -> %s", email.Subject, toEmail)
	return nil
}

// sendMailWithFallback 尝试多种方式发送邮件
func (c *MailClient) sendMailWithFallback(smtpServer, smtpPort, toEmail string, body []byte) error {
	// 方法1: 尝试 STARTTLS (端口587)
	if smtpPort == "587" {
		err := c.sendMailWithSTARTTLS(smtpServer, smtpPort, toEmail, body)
		if err == nil {
			return nil
		}
	}

	// 方法2: 尝试 SSL/TLS (端口465)
	err := c.sendMailWithSSL(smtpServer, "465", toEmail, body)
	if err == nil {
		return nil
	}

	// 方法3: 尝试普通连接 (端口25)
	err = c.sendMailPlain(smtpServer, "25", toEmail, body)
	if err == nil {
		return nil
	}

	return fmt.Errorf("所有发送方式都失败")
}

// sendMailWithSTARTTLS 使用 STARTTLS 发送邮件
func (c *MailClient) sendMailWithSTARTTLS(smtpServer, port, toEmail string, body []byte) error {
	addr := fmt.Sprintf("%s:%s", smtpServer, port)

	// 连接到SMTP服务器
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %v", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %v", err)
	}
	defer client.Close()

	// 启用STARTTLS
	if err = client.StartTLS(&tls.Config{ServerName: smtpServer}); err != nil {
		return fmt.Errorf("启用STARTTLS失败: %v", err)
	}

	// 认证
	if err = client.Auth(smtp.PlainAuth("", c.config.Username, c.config.Password, smtpServer)); err != nil {
		return fmt.Errorf("SMTP认证失败: %v", err)
	}

	// 发送邮件
	if err = client.Mail(c.config.Username); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("设置收件人失败: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("开始发送邮件数据失败: %v", err)
	}

	_, err = w.Write(body)
	if err != nil {
		return fmt.Errorf("写入邮件数据失败: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("完成发送邮件失败: %v", err)
	}

	return nil
}

// sendMailWithSSL 使用 SSL/TLS 发送邮件
func (c *MailClient) sendMailWithSSL(smtpServer, port, toEmail string, body []byte) error {
	addr := fmt.Sprintf("%s:%s", smtpServer, port)

	// 创建TLS配置
	tlsConfig := &tls.Config{
		ServerName: smtpServer,
	}

	// 连接到SMTP服务器
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %v", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %v", err)
	}
	defer client.Close()

	// 认证
	if err = client.Auth(smtp.PlainAuth("", c.config.Username, c.config.Password, smtpServer)); err != nil {
		return fmt.Errorf("SMTP认证失败: %v", err)
	}

	// 发送邮件
	if err = client.Mail(c.config.Username); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("设置收件人失败: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("开始发送邮件数据失败: %v", err)
	}

	_, err = w.Write(body)
	if err != nil {
		return fmt.Errorf("写入邮件数据失败: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("完成发送邮件失败: %v", err)
	}

	return nil
}

// sendMailPlain 使用普通连接发送邮件
func (c *MailClient) sendMailPlain(smtpServer, port, toEmail string, body []byte) error {
	addr := fmt.Sprintf("%s:%s", smtpServer, port)

	// 连接到SMTP服务器
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %v", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %v", err)
	}
	defer client.Close()

	// 发送邮件
	if err = client.Mail(c.config.Username); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("设置收件人失败: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("开始发送邮件数据失败: %v", err)
	}

	_, err = w.Write(body)
	if err != nil {
		return fmt.Errorf("写入邮件数据失败: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("完成发送邮件失败: %v", err)
	}

	return nil
}

// SendRawEmail 发送原始邮件数据
func (c *MailClient) SendRawEmail(rawData []byte, toEmail string) error {
	// 添加转发头信息
	forwardedData := c.addForwardHeaders(rawData, toEmail)

	// 解析SMTP服务器地址和端口
	smtpServer := c.getSMTPServer()
	smtpPort := c.getSMTPPort()

	// 尝试发送邮件
	err := c.sendMailWithFallback(smtpServer, smtpPort, toEmail, forwardedData)
	if err != nil {
		return fmt.Errorf("发送原始邮件失败: %v", err)
	}

	log.Printf("IMAP Provider 原始邮件发送成功: %s", toEmail)
	return nil
}

// addForwardHeaders 添加转发头信息
func (c *MailClient) addForwardHeaders(rawData []byte, toEmail string) []byte {
	// 添加转发相关的头信息
	headers := []string{
		fmt.Sprintf("Resent-From: %s", c.config.Username),
		fmt.Sprintf("Resent-To: %s", toEmail),
		"X-Forwarded-By: Mail-Dispatcher-System",
		"",
	}

	// 在原始数据前添加头信息
	var result []byte
	for _, header := range headers {
		result = append(result, []byte(header+"\r\n")...)
	}
	result = append(result, rawData...)

	return result
}

// getSMTPServer 获取SMTP服务器地址
func (c *MailClient) getSMTPServer() string {
	// 根据IMAP服务器推断SMTP服务器
	server := c.config.Server
	if strings.Contains(server, "qq.com") {
		return "smtp.qq.com"
	} else if strings.Contains(server, "gmail.com") {
		return "smtp.gmail.com"
	} else if strings.Contains(server, "163.com") {
		return "smtp.163.com"
	} else if strings.Contains(server, "126.com") {
		return "smtp.126.com"
	}
	// 默认使用IMAP服务器对应的SMTP服务器
	return strings.Replace(server, "imap.", "smtp.", 1)
}

// getSMTPPort 获取SMTP端口
func (c *MailClient) getSMTPPort() string {
	// 根据IMAP服务器推断SMTP端口
	server := c.config.Server
	if strings.Contains(server, "qq.com") {
		return "587"
	} else if strings.Contains(server, "gmail.com") {
		return "587"
	} else if strings.Contains(server, "163.com") {
		return "25"
	} else if strings.Contains(server, "126.com") {
		return "25"
	}
	return "587" // 默认使用587端口
}

// parseMessage 解析IMAP消息
func (c *MailClient) parseMessage(msg *imap.Message) (models.Email, error) {
	email := models.Email{
		ReceivedAt: time.Now(),
	}

	// 首先尝试使用 IMAP envelope 数据
	if msg.Envelope != nil {
		if len(msg.Envelope.Subject) > 0 {
			// 获取完整的主题字节数组
			subjectBytes := msg.Envelope.Subject
			// 直接使用 UTF-8 解码
			subject := string(subjectBytes)
			email.Subject = subject
		}
		if len(msg.Envelope.From) > 0 {
			email.From = msg.Envelope.From[0].Address()
		}
		if len(msg.Envelope.To) > 0 {
			email.To = msg.Envelope.To[0].Address()
		}
	}

	// 如果 envelope 没有数据，尝试使用 go-message 解析邮件正文
	if email.Subject == "" && msg.Body != nil {
		section := &imap.BodySectionName{}
		r := msg.GetBody(section)
		if r != nil {
			// 使用 go-message 解析邮件
			entity, err := message.Read(r)
			if err != nil {
				log.Printf("解析邮件失败: %v", err)
			} else {
				// 解析邮件头
				header := entity.Header
				if subject := header.Get("Subject"); subject != "" {
					email.Subject = subject
				}
				if from := header.Get("From"); from != "" {
					email.From = from
				}
				if to := header.Get("To"); to != "" {
					email.To = to
				}
			}
		}
	}

	// 生成MessageID
	email.MessageID = fmt.Sprintf("%d-%d", c.config.AccountID, msg.Uid)

	return email, nil
}
