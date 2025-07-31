# Mail Forwarding System

A Golang-based mail forwarding system that supports fetching emails from multiple email accounts and automatically forwarding them to target emails based on subject keywords.

## Features

- **Multi-email Support**: Supports Gmail, QQ, Outlook and other IMAP-compatible email services
- **Smart Forwarding**: Automatically forwards emails based on subject format "Keyword - Target Name"
- **Unified Mail Client**: Uses MailClient to handle both email fetching and sending
- **RESTful API**: Provides complete APIs for account, target, and log management
- **Real-time Polling**: Timed polling to fetch new emails, ensuring timely processing

## Quick Start

### 1. Requirements

- Go 1.24.5+
- MySQL 8.0+

### 2. Using Docker Compose

```bash
# Start MySQL database
docker-compose up -d mysql

# Start application
docker-compose up -d
```

### 3. Local Development

```bash
# Install dependencies
go mod tidy

# Start application
go run cmd/main.go
```

The application will start at `http://localhost:8080`.

## API Endpoints

### Forward Target Management

- `GET /api/v1/targets` - Get all forward targets
- `POST /api/v1/targets` - Create forward target
- `PUT /api/v1/targets/:id` - Update forward target
- `DELETE /api/v1/targets/:id` - Delete forward target

### Email Account Management

- `GET /api/v1/accounts` - Get all email accounts
- `POST /api/v1/accounts` - Create email account
- `PUT /api/v1/accounts/:id` - Update email account
- `DELETE /api/v1/accounts/:id` - Delete email account
- `PUT /api/v1/accounts/:id/toggle` - Toggle account status

### Mail Log Management

- `GET /api/v1/logs` - Get mail logs
- `GET /api/v1/logs/failed` - Get failed logs
- `GET /api/v1/logs/successful` - Get successful logs
- `GET /api/v1/logs/stats` - Get log statistics

## Usage Examples

### 1. Create Forward Target

```bash
curl -X POST http://localhost:8080/api/v1/targets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "description": "Technical Manager"
  }'
```

### 2. Add Email Account

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

### 3. Mail Forwarding Rules

The system forwards emails based on subject format: `Keyword - Target Name`

Examples:
- Subject: `Alert - John Doe` → Forward to `john@example.com`
- Subject: `Notification - Finance` → Forward to `finance@company.com`

## Configuration

### Database Configuration

The system uses MySQL 8.0, automatically configured via Docker Compose:

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

### Email Configuration

#### Gmail Configuration
1. Enable two-factor authentication
2. Generate app-specific password
3. Use app-specific password as password field

#### QQ Email Configuration
1. Enable IMAP service
2. Use authorization code as password

### Application Configuration

System configuration is managed through environment variables:

```bash
# Server configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=mail_dispatcher
DB_PASSWORD=mail_dispatcher_password
DB_NAME=mail_dispatcher
DB_CHARSET=utf8mb4

# Mail configuration
MAIL_POLLING_INTERVAL=300
MAIL_MAX_RETRY_COUNT=3
MAIL_RETRY_INTERVAL=60
```

## Project Structure

```
mail-dispatcher/
├── cmd/main.go                    # Main program entry
├── internal/
│   ├── config/                    # Configuration management
│   ├── controllers/               # HTTP controllers
│   ├── mail/                      # Mail client
│   ├── models/                    # Data models
│   ├── routes/                    # Route definitions
│   └── services/                  # Business services

├── docker-compose.yml             # Docker orchestration
└── README.md                      # Project documentation
```

## Testing

```bash
# Run all tests
go test ./...

# Run mail client tests
go test ./internal/mail -v

# Run service tests
go test ./internal/services -v
```

## Troubleshooting

### Common Issues

1. **IMAP Connection Failed**
   - Check if email account password is correct
   - Confirm email provider supports IMAP
   - Check network connection

2. **Email Duplicate Forwarding**
   - Check email log records in database
   - Confirm email Message-ID uniqueness

3. **MySQL Connection Failed**
   - Confirm MySQL service is running
   - Check database configuration

### Log Viewing

```bash
# View failed processing records
curl "http://localhost:8080/api/v1/logs/failed?limit=10"

# View specific account processing records
curl "http://localhost:8080/api/v1/logs?account_id=1&limit=10"
```

## Development

### Core Concepts

- **MailClient**: Unified mail client supporting IMAP fetching and SMTP sending
- **Dynamic Creation**: MailClient is dynamically created for each polling cycle, ensuring latest configuration
- **Polling Mechanism**: Timed polling to fetch new emails, avoiding complex real-time push
- **Subject Parsing**: Automatically match forward targets based on email subject format

### Extension Development

To add new email support:

1. Ensure email supports IMAP protocol
2. Add corresponding server configuration in MailClient
3. Add new email account via API

The system automatically handles email fetching and forwarding logic.