package models

import (
	"time"

	"gorm.io/gorm"
)

// ForwardTarget 转发目标表
type ForwardTarget struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex;size:100;not null;comment:转发对象名称"`
	Email       string `gorm:"size:255;not null;comment:目标邮箱地址"`
	Description string `gorm:"size:500;comment:描述或备注"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// MailAccount 邮箱账户表
type MailAccount struct {
	ID        uint   `gorm:"primaryKey"`
	Address   string `gorm:"size:255;not null;comment:邮箱地址"`
	Username  string `gorm:"size:255;comment:登录用户名"`
	Password  string `gorm:"size:500;comment:密码或OAuth token"`
	Server    string `gorm:"size:255;comment:IMAP服务器地址"`
	Settings  string `gorm:"type:text;comment:其他配置JSON"`
	LastUID   uint32 `gorm:"comment:上次处理的IMAP UID"`
	IsActive  bool   `gorm:"default:true;comment:是否启用"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// MailLog 邮件处理日志表
type MailLog struct {
	ID          uint        `gorm:"primaryKey"`
	AccountID   uint        `gorm:"not null;comment:来源账户ID"`
	Account     MailAccount `gorm:"foreignKey:AccountID"`
	MessageID   string      `gorm:"size:500;comment:邮件Message-ID"`
	Subject     string      `gorm:"size:500;comment:邮件主题"`
	From        string      `gorm:"size:255;comment:发件人地址"`
	To          string      `gorm:"size:255;comment:原邮件收件人"`
	ReceivedAt  time.Time   `gorm:"comment:邮件接收时间"`
	ForwardTo   string      `gorm:"size:255;comment:转发目标地址"`
	Status      string      `gorm:"size:50;not null;comment:处理状态"`
	Error       string      `gorm:"type:text;comment:错误信息"`
	ForwardedAt *time.Time  `gorm:"comment:转发时间"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Email 内部邮件结构
type Email struct {
	MessageID  string
	Subject    string
	From       string
	To         string
	Body       string
	RawData    []byte
	ReceivedAt time.Time
}
