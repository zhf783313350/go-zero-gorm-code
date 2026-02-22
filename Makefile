.PHONY: build run test clean docker docker-push deploy-k8s

# 变量
APP_NAME := access-control
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := latest
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# 构建
build:
	go build -ldflags="-X main.Version=$(VERSION)" -o bin/$(APP_NAME) .

# 生产构建
build-prod:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-w -s -X main.Version=$(VERSION)" \
		-o bin/$(APP_NAME) .

# 运行
run:
	go run main.go -f etc/config.yaml

# 测试
test:
	go test -v -race -cover ./...

# 测试覆盖率报告
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 代码检查
lint:
	golangci-lint run ./...

# 清理
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

# 下载依赖
tidy:
	go mod tidy

# ==================== Docker 相关 ====================

# Docker 构建 (开发环境)
docker:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Docker 构建 (生产环境)
docker-prod:
	docker build -f Dockerfile.prod -t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--build-arg VERSION=$(VERSION) .

# Docker Compose 启动 (开发环境)
docker-up:
	docker-compose up -d

# Docker Compose 启动 (生产环境)
docker-up-prod:
	docker-compose -f docker-compose.prod.yml up -d

# Docker Compose 停止
docker-down:
	docker-compose down

# Docker Compose 停止 (生产环境)
docker-down-prod:
	docker-compose -f docker-compose.prod.yml down

# Docker Compose 日志
docker-logs:
	docker-compose logs -f

# Docker Compose 构建并启动 (生产环境)
docker-deploy:
	docker-compose -f docker-compose.prod.yml up -d --build

# ==================== K8s 相关 ====================

# K8s 部署
deploy-k8s:
	kubectl apply -f deploy/k8s/namespace.yaml
	kubectl apply -f deploy/k8s/rbac.yaml
	kubectl apply -f deploy/k8s/secret.yaml
	kubectl apply -f deploy/k8s/configmap.yaml
	kubectl apply -f deploy/k8s/deployment.yaml
	kubectl apply -f deploy/k8s/service.yaml
	kubectl apply -f deploy/k8s/ingress.yaml
	kubectl apply -f deploy/k8s/hpa.yaml
	kubectl apply -f deploy/k8s/pdb.yaml

# K8s 删除
undeploy-k8s:
	kubectl delete -f deploy/k8s/ --ignore-not-found

# K8s 状态
status-k8s:
	kubectl get all -n access-control

# K8s 重启
restart-k8s:
	kubectl rollout restart deployment/access-control-api -n access-control

# K8s 日志
logs-k8s:
	kubectl logs -f -l app=access-control-api -n access-control

# ==================== 数据库相关 ====================

# 数据库备份
db-backup:
	@mkdir -p backups
	docker-compose -f docker-compose.prod.yml exec -T postgres \
		pg_dump -U postgres access_control > backups/backup_$$(date +%Y%m%d_%H%M%S).sql

# 数据库迁移 (使用 golang-migrate CLI, 如果本地安装了的话)
# 也可以直接运行应用，应用启动时会自动 migrate
migrate-create:
	@read -p "请输入迁移名称: " name; \
	migrate create -ext sql -dir internal/svc/migrations -seq $$name

# 生成 API 文档
swagger:
	swag init

# 帮助
help:
	@echo "==================== 构建相关 ===================="
	@echo "  build          - 构建应用"
	@echo "  build-prod     - 生产环境构建"
	@echo "  run            - 运行应用 (包含自动迁移)"
	@echo "  test           - 运行测试"
	@echo "  test-coverage  - 生成测试覆盖率报告"
	@echo "  lint           - 代码检查"
	@echo "  clean          - 清理构建产物"
	@echo "  tidy           - 下载依赖"
	@echo ""
	@echo "==================== Docker 相关 ===================="
	@echo "  docker         - 构建 Docker 镜像 (开发)"
	@echo "  docker-prod    - 构建 Docker 镜像 (生产)"
	@echo "  docker-up      - 启动 (开发环境)"
	@echo "  docker-up-prod - 启动 (生产环境)"
	@echo "  docker-down    - 停止"
	@echo "  docker-deploy  - 构建并部署 (生产)"
	@echo "  docker-logs    - 查看日志"
	@echo ""
	@echo "==================== K8s 相关 ===================="
	@echo "  deploy-k8s     - 部署到 Kubernetes"
	@echo "  undeploy-k8s   - 从 Kubernetes 删除"
	@echo "  status-k8s     - 查看 K8s 状态"
	@echo "  restart-k8s    - 重启 K8s 部署"
	@echo "  logs-k8s       - 查看 K8s 日志"
	@echo ""
	@echo "==================== 数据库相关 ===================="
	@echo "  db-backup      - 备份数据库"
	@echo "  migrate-create - 创建新的数据库迁移脚本"
	@echo ""
