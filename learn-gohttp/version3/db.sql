-- 使用 webapp_db 数据库
USE webapp_db;

-- 如果 users 表已存在，则删除它
DROP TABLE IF EXISTS users;

-- 重新创建 users 表，增加 password_hash 字段
CREATE TABLE users (
    name VARCHAR(50) PRIMARY KEY,
    age INT,
    city VARCHAR(50),
    password_hash VARCHAR(255) NOT NULL -- 用于存储 bcrypt 哈希后的密码
);

-- 如果 products 表已存在，则删除它
DROP TABLE IF EXISTS products;

-- 创建 products 表
CREATE TABLE products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入一些商品样本数据
INSERT INTO products (name, price) VALUES
('笔记本电脑', 6999.00),
('机械键盘', 899.00),
('显示器', 1599.50);

