package mail

import (
	"mail-dispatcher/internal/config"
	"testing"
)

func TestMailClient_Init(t *testing.T) {
	cfg := &config.Config{}
	client := NewMailClient(cfg)
	config := Config{
		AccountID: 1,
		Provider:  "mail",
		Address:   "test@example.com",
		Username:  "test@example.com",
		Password:  "password",
		Server:    "imap.example.com:993",
	}

	err := client.Init(config)
	if err == nil {
		t.Error("应该失败，因为没有真实的IMAP服务器")
	}
}

func TestMailClient_GetName(t *testing.T) {
	cfg := &config.Config{}
	client := NewMailClient(cfg)
	if client.GetName() != "IMAP" {
		t.Errorf("期望 'IMAP'，得到 '%s'", client.GetName())
	}
}
