-- 邮件转发系统数据库初始化脚本

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS mail_dispatcher CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 设置 MySQL 8.0 兼容性
SET GLOBAL sql_mode = 'STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO';

-- 使用数据库
USE mail_dispatcher;