package mail

// Config 邮件客户端配置
type Config struct {
	AccountID uint
	Provider  string
	Address   string
	Username  string
	Password  string
	Server    string
	Settings  string
	LastUID   uint32
}
