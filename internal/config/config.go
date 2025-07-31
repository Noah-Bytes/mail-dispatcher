package config

import (
	"os"
	"strconv"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Mail     MailConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Charset  string
}

// MailConfig 邮件处理配置
type MailConfig struct {
	PollingInterval int
	MaxRetryCount   int
	RetryInterval   int
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "mail_dispatcher"),
			Password: getEnv("DB_PASSWORD", "mail_dispatcher_password"),
			DBName:   getEnv("DB_NAME", "mail_dispatcher"),
			Charset:  getEnv("DB_CHARSET", "utf8mb4"),
		},
		Mail: MailConfig{
			PollingInterval: getEnvInt("MAIL_POLLING_INTERVAL", 300),
			MaxRetryCount:   getEnvInt("MAIL_MAX_RETRY_COUNT", 3),
			RetryInterval:   getEnvInt("MAIL_RETRY_INTERVAL", 60),
		},
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return c.Database.User + ":" + c.Database.Password + "@tcp(" +
		c.Database.Host + ":" + c.Database.Port + ")/" +
		c.Database.DBName + "?charset=" + c.Database.Charset +
		"&parseTime=True&loc=Local"
}
