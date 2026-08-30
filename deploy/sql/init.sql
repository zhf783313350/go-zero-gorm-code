-- 与应用迁移保持一致的基础数据结构。
-- 正式环境由 internal/svc/migrations 执行版本迁移；此文件仅用于初始化数据库。
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    "phoneNumber" VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255),
    status INTEGER DEFAULT 1,
    "validTime" VARCHAR(50),
    "type" VARCHAR(50)
);

CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users("phoneNumber");
