# 邮件转发系统 Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop install dev

# 默认目标
help:
	@echo "邮件转发系统构建和部署工具"
	@echo ""
	@echo "可用命令:"
	@echo "  build        - 构建应用"
	@echo "  run          - 运行应用"
	@echo "  test         - 运行所有测试"
	@echo "  test-mail    - 运行邮件客户端测试"
	@echo "  test-services- 运行服务测试"
	@echo "  test-coverage- 运行测试并生成覆盖率报告"
	@echo "  clean        - 清理构建文件"
	@echo "  install      - 安装依赖"
	@echo "  docker-build - 构建 Docker 镜像"
	@echo "  docker-run   - 使用 Docker Compose 启动服务"
	@echo "  docker-stop  - 停止 Docker 服务"
	@echo "  docker-logs  - 查看服务日志"
	@echo "  dev          - 开发模式（热重载）"
	@echo "  api-test     - 测试 API 接口"

# 构建应用
build:
	@echo "构建邮件转发系统..."
	go build -o mail-dispatcher cmd/main.go
	@echo "构建完成: mail-dispatcher"

# 运行应用
run: build
	@echo "启动邮件转发系统..."
	./mail-dispatcher

# 运行所有测试
test:
	@echo "运行所有测试..."
	go test -v ./...

# 运行邮件客户端测试
test-mail:
	@echo "运行邮件客户端测试..."
	go test -v ./internal/mail

# 运行服务测试
test-services:
	@echo "运行服务测试..."
	go test -v ./internal/services

# 运行控制器测试
test-controllers:
	@echo "运行控制器测试..."
	go test -v ./internal/controllers

# 测试覆盖率
test-coverage:
	@echo "运行测试并生成覆盖率报告..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 测试 API 接口
api-test:
	@echo "测试 API 接口..."
	@if [ -f scripts/test_api.sh ]; then \
		chmod +x scripts/test_api.sh; \
		./scripts/test_api.sh; \
	else \
		echo "API 测试脚本不存在: scripts/test_api.sh"; \
	fi

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -f mail-dispatcher
	rm -f coverage.out coverage.html
	@echo "清理完成"

# 安装依赖
install:
	@echo "安装依赖..."
	go mod tidy
	go mod download
	@echo "依赖安装完成"

# 构建 Docker 镜像
docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t mail-dispatcher .
	@echo "Docker 镜像构建完成"

# 使用 Docker Compose 启动服务
docker-run:
	@echo "启动 Docker 服务..."
	docker compose up -d
	@echo "服务启动完成，访问 http://localhost:8080"

# 停止 Docker 服务
docker-stop:
	@echo "停止 Docker 服务..."
	docker compose down
	@echo "服务已停止"

# 查看服务状态
docker-status:
	@echo "查看服务状态..."
	docker compose ps

# 查看服务日志
docker-logs:
	@echo "查看服务日志..."
	docker compose logs -f

# 重新构建并启动服务
docker-rebuild: docker-stop docker-build docker-run

# 开发模式（热重载）
dev:
	@echo "启动开发模式..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "请先安装 air: go install github.com/cosmtrek/air@latest"; \
		echo "使用 go run 启动..."; \
		go run cmd/main.go; \
	fi

# 快速开发（直接运行）
dev-run:
	@echo "快速开发模式..."
	go run cmd/main.go

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "请先安装 golangci-lint"; \
		echo "安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 生成文档
docs:
	@echo "生成 API 文档..."
	@echo "API 文档请参考 README.md 和 README.zh.md"

# 数据库初始化
db-init:
	@echo "初始化数据库..."
	@if [ -f init.sql ]; then \
		echo "使用 init.sql 初始化数据库..."; \
		mysql -u mail_dispatcher -p mail_dispatcher < init.sql; \
	else \
		echo "init.sql 文件不存在"; \
	fi

# 完整测试套件
test-all: test test-mail test-services test-controllers test-coverage
	@echo "所有测试完成"

# 生产环境构建
prod-build:
	@echo "生产环境构建..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mail-dispatcher cmd/main.go
	@echo "生产环境构建完成"

# 检查依赖更新
deps-check:
	@echo "检查依赖更新..."
	go list -u -m all 