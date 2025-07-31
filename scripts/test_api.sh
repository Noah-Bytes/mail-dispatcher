#!/bin/bash

# API 测试脚本
# 使用方法: ./scripts/test_api.sh

BASE_URL="http://localhost:8080"

echo "开始测试邮件转发系统 API..."

# 测试健康检查
echo "1. 测试健康检查..."
curl -s "$BASE_URL/ping" | jq .

# 测试获取转发目标
echo -e "\n2. 测试获取转发目标..."
curl -s "$BASE_URL/api/v1/targets" | jq .

# 测试创建转发目标
echo -e "\n3. 测试创建转发目标..."
curl -s -X POST "$BASE_URL/api/v1/targets" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试用户",
    "email": "test@example.com",
    "description": "测试用户"
  }' | jq .

# 测试获取邮箱账户
echo -e "\n4. 测试获取邮箱账户..."
curl -s "$BASE_URL/api/v1/accounts" | jq .

# 测试创建邮箱账户
echo -e "\n5. 测试创建邮箱账户..."
curl -s -X POST "$BASE_URL/api/v1/accounts" \
  -H "Content-Type: application/json" \
  -d '{
    "address": "test@gmail.com",
    "username": "test@gmail.com",
    "password": "test_password",
    "server": "imap.gmail.com:993"
  }' | jq .

# 测试获取日志统计
echo -e "\n6. 测试获取日志统计..."
curl -s "$BASE_URL/api/v1/logs/stats" | jq .

# 测试获取失败日志
echo -e "\n7. 测试获取失败日志..."
curl -s "$BASE_URL/api/v1/logs/failed?limit=5" | jq .

echo -e "\nAPI 测试完成！" 