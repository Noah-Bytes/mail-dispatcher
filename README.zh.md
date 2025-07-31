# 邮件转发系统

一个基于 Golang 的邮件转发系统，支持从多个邮箱账户获取邮件，根据主题关键词自动转发到目标邮箱。

## 系统特性

- **多邮箱支持**: 支持 Gmail、QQ、Outlook 等支持 IMAP 的邮箱
- **智能转发**: 根据邮件主题格式 "关键词 - 目标名称" 自动转发
- **统一邮件客户端**: 使用 MailClient 统一处理邮件获取和发送
- **RESTful API**: 提供完整的账户、目标、日志管理接口
- **实时轮询**: 定时轮询获取新邮件，确保及时处理

## 快速开始

### 1. 环境要求

- Go 1.24.5+
- MySQL 8.0+

### 2. 使用 Docker Compose 启动

```bash
# 启动 MySQL 数据库
docker-compose up -d mysql

# 启动应用
docker-compose up -d
```

### 3. 本地开发

```bash
# 安装依赖
go mod tidy

# 启动应用
go run cmd/main.go
```

应用将在 `http://localhost:8080` 启动。

## API 接口

### 转发目标管理

- `GET /api/v1/targets` - 获取所有转发目标
- `POST /api/v1/targets` - 创建转发目标
- `PUT /api/v1/targets/:id` - 更新转发目标
- `DELETE /api/v1/targets/:id` - 删除转发目标

### 邮箱账户管理

- `GET /api/v1/accounts` - 获取所有邮箱账户
- `POST /api/v1/accounts` - 创建邮箱账户
- `PUT /api/v1/accounts/:id` - 更新邮箱账户
- `DELETE /api/v1/accounts/:id` - 删除邮箱账户
- `PUT /api/v1/accounts/:id/toggle` - 切换账户状态

### 邮件日志管理

- `GET /api/v1/logs` - 获取邮件日志
- `GET /api/v1/logs/failed` - 获取失败的日志
- `GET /api/v1/logs/successful` - 获取成功的日志
- `GET /api/v1/logs/stats` - 获取日志统计信息

## 使用示例

### 1. 创建转发目标

```bash
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三",
    "email": "zhangsan@example.com",
    "description": "技术部经理"
  }'
```

### 2. 添加邮箱账户

```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "address": "router@gmail.com",
    "username": "router@gmail.com",
    "password": "your_app_password",
    "server": "imap.gmail.com:993"
  }'
```

### 3. 邮件转发规则

系统根据邮件主题进行转发，主题格式为：`关键词 - 目标名称`

例如：
- 主题：`报警 - 张三` → 转发给 `zhangsan@example.com`
- 主题：`通知 - 财务部` → 转发给 `finance@company.com`

## 配置说明

### 数据库配置

系统使用 MySQL 8.0，通过 Docker Compose 自动配置：

```yaml
# docker-compose.yml
mysql:
  image: mysql:8.0
  environment:
    MYSQL_ROOT_PASSWORD: root_password
    MYSQL_DATABASE: mail_dispatcher
    MYSQL_USER: mail_dispatcher
    MYSQL_PASSWORD: mail_dispatcher_password
```

### 邮箱配置

#### Gmail 配置
1. 开启两步验证
2. 生成应用专用密码
3. 使用应用专用密码作为密码字段

#### QQ 邮箱配置
1. 开启 IMAP 服务
2. 使用授权码作为密码

### 应用配置

系统配置通过环境变量管理：

```bash
# 服务器配置
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=mail_dispatcher
DB_PASSWORD=mail_dispatcher_password
DB_NAME=mail_dispatcher
DB_CHARSET=utf8mb4

# 邮件配置
MAIL_POLLING_INTERVAL=300
MAIL_MAX_RETRY_COUNT=3
MAIL_RETRY_INTERVAL=60
```

## 项目结构

```
mail-dispatcher/
├── cmd/main.go                    # 主程序入口
├── internal/
│   ├── config/                    # 配置管理
│   ├── controllers/               # HTTP 控制器
│   ├── mail/                      # 邮件客户端
│   ├── models/                    # 数据模型
│   ├── routes/                    # 路由定义
│   └── services/                  # 业务服务

├── docker-compose.yml             # Docker 编排
└── README.md                      # 项目说明
```

## 测试

```bash
# 运行所有测试
go test ./...

# 运行邮件客户端测试
go test ./internal/mail -v

# 运行服务测试
go test ./internal/services -v
```

## 故障排除

### 常见问题

1. **IMAP 连接失败**
   - 检查邮箱账户密码是否正确
   - 确认邮箱服务商是否支持 IMAP
   - 检查网络连接

2. **邮件重复转发**
   - 检查数据库中的邮件日志记录
   - 确认邮件 Message-ID 的唯一性

3. **MySQL 连接失败**
   - 确认 MySQL 服务已启动
   - 检查数据库配置信息

### 日志查看

```bash
# 查看失败的处理记录
curl "http://localhost:8080/api/v1/logs/failed?limit=10"

# 查看特定账户的处理记录
curl "http://localhost:8080/api/v1/logs?account_id=1&limit=10"
```

## 开发说明

### 核心概念

- **MailClient**: 统一的邮件客户端，支持 IMAP 获取和 SMTP 发送
- **动态创建**: 每次轮询时动态创建 MailClient，确保配置最新
- **轮询机制**: 定时轮询获取新邮件，避免复杂的实时推送
- **主题解析**: 根据邮件主题格式自动匹配转发目标

### 扩展开发

要添加新的邮箱支持，只需要：

1. 确保邮箱支持 IMAP 协议
2. 在 MailClient 中添加对应的服务器配置
3. 通过 API 添加新的邮箱账户

系统会自动处理邮件获取和转发逻辑。 