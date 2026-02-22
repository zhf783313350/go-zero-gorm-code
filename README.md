# Access Control 权限管理系统

基于 go-zero + Docker + Kubernetes 的微服务权限管理系统

## 🚀 特性

- **高性能**: 基于 Go-Zero 框架，支持高并发
- **自动化迁移**: 集成 `golang-migrate`，程序启动自动同步数据库表结构
- **容器化**: 完整的 Docker 支持，一键部署
- **K8s 原生**: 支持 Kubernetes 部署，自动扩缩容
- **监控完善**: 集成 Prometheus + Grafana 监控
- **负载均衡**: Nginx 反向代理，自动负载均衡
- **安全可靠**: JWT 认证，接口限流 (Token Bucket)
- **并发保护**: 使用 `SingleFlight` 防御缓存击穿，完善的缓存穿透保护

## 📁 项目结构

```
access-control/
├── main.go                      # 程序入口
├── etc/
│   ├── config.yaml             # 本地开发配置
│   ├── config-docker.yaml      # Docker 配置
│   └── config-prod.yaml        # 生产环境配置
├── internal/
│   ├── config/                 # 配置结构体
│   ├── handler/                # HTTP 处理器
│   ├── logic/                  # 业务逻辑
│   ├── middleware/             # 中间件
│   ├── model/                  # 数据模型
│   ├── svc/                    # 服务上下文
│   │   └── migrations/         # 数据库版本迁移脚本 [NEW]
│   └── types/                  # 类型定义
├── deploy/
│   ├── k8s/                    # Kubernetes 部署文件
│   ├── nginx/                  # Nginx 配置
│   ├── prometheus/             # Prometheus 配置
│   └── grafana/                # Grafana 配置
├── Dockerfile                   # Docker 构建文件 (Go 1.26.0)
├── Dockerfile.prod             # Docker 构建文件 (生产环境，Go 1.26.0)
├── docker-compose.yml          # Docker Compose (开发)
├── docker-compose.prod.yml     # Docker Compose (生产)
├── deploy.sh                   # 部署脚本 (Linux/Mac)
├── Makefile                    # 构建脚本 (包含迁移管理)
└── README.md
```

## 🛠 快速开始

### 环境要求

- **Go 1.26.0+** [UPDATED]
- Docker 20.0+
- Docker Compose 2.0+
- PostgreSQL 14+ (或使用 Docker)
- Redis 6+ (或使用 Docker)

### 本地开发

```bash
# 1. 克隆项目
git clone https://github.com/zhf783313350/access-control.git
cd access-control

# 2. 安装依赖
go mod tidy

# 3. 启动数据库和 Redis (使用 Docker)
docker-compose up -d postgres redis

# 4. 运行应用 (程序会自动运行数据库迁移)
make run
```

### Docker 部署 (开发环境)

```bash
# 构建并启动
docker-compose up -d --build

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### Docker 部署 (生产环境)

```bash
# 1. 复制环境变量配置
cp .env.example .env

# 2. 修改 .env 文件中的配置 (必须修改 JWT_SECRET 等)

# 3. 启动所有服务
./deploy.sh start
```

## 🔧 常用命令

### 数据库管理 (golang-migrate)

```bash
# 创建新的数据库迁移脚本
make migrate-create
```

### 其他常用命令

```bash
make help           # 查看所有命令
make build          # 构建应用
make test           # 运行测试
make lint           # 代码检查
make docker-prod    # 构建生产镜像
make db-backup      # 备份数据库
```

## 📡 API 接口 (部分说明)

### 1. 查询用户

`POST /api/user/query` (根据手机号查询，带 SingleFlight 保护)

### 2. 编辑用户

`POST /api/user/edit` (根据手机号标识进行更新)

```json
{
  "phoneNumber": "13800000001",
  "status": 1,
  "validTime": "2026-12-31 23:59:59"
}
```

## 📈 性能与稳定性

- **缓存击穿**: 应用层 `SingleFlight` 确保同一个 Key 只有一个请求穿透到 DB。
- **缓存穿透**: 对不存在的 Key 缓存 `empty` 占位符。
- **限流保护**: 集成分布式令牌桶限流，防止瞬时峰值。
- **优雅部署**: Docker 多阶段构建 + K8s HPA/PDB 支持。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 License

MIT License
