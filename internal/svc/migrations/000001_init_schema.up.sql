-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    "phoneNumber" VARCHAR(50) UNIQUE NOT NULL,
    status INTEGER DEFAULT 1,
    "validTime" VARCHAR(50)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users("phoneNumber");

-- 插入初始测试数据 (可选，通常建议放在单独的 seed 文件或由逻辑生成，但这里为了保持 init.sql 的功能先保留)
INSERT INTO users ("phoneNumber", status, "validTime") VALUES
('13800000001', 1, '2026-12-31 23:59:59'),
('13800000002', 1, '2026-12-31 23:59:59'),
('13800000003', 1, '2026-12-31 23:59:59'),
('13800000004', 1, '2026-12-31 23:59:59'),
('13800000005', 1, '2026-12-31 23:59:59')
ON CONFLICT ("phoneNumber") DO NOTHING;
