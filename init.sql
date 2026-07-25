-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    "phoneNumber" VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255),
    status INTEGER DEFAULT 1,
    "validTime" VARCHAR(50)
);

-- 插入测试数据 (密码: 123456)
INSERT INTO users ("phoneNumber", password, status, "validTime") VALUES
('13800000001', '$2a$10$wpmdLxGvNSjhHHVq7JurlOEfE2hWxRR.mndC8NDP6K5HKZxIJk606', 1, '2026-12-31 23:59:59'),
('13800000002', '$2a$10$wpmdLxGvNSjhHHVq7JurlOEfE2hWxRR.mndC8NDP6K5HKZxIJk606', 1, '2026-12-31 23:59:59'),
('13800000003', '$2a$10$wpmdLxGvNSjhHHVq7JurlOEfE2hWxRR.mndC8NDP6K5HKZxIJk606', 1, '2026-12-31 23:59:59'),
('13800000004', '$2a$10$wpmdLxGvNSjhHHVq7JurlOEfE2hWxRR.mndC8NDP6K5HKZxIJk606', 1, '2026-12-31 23:59:59'),
('13800000005', '$2a$10$wpmdLxGvNSjhHHVq7JurlOEfE2hWxRR.mndC8NDP6K5HKZxIJk606', 1, '2026-12-31 23:59:59')
ON CONFLICT ("phoneNumber") DO NOTHING;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users("phoneNumber");
