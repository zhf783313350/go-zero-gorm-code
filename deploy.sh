#!/bin/bash
# 生产环境部署脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 配置
PROJECT_NAME="go-zero-gorm-code"
COMPOSE_FILE="docker-compose.prod.yml"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  go-zero-gorm-code 生产环境部署脚本${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查 .env 文件
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}警告: .env 文件不存在，将从 .env.example 复制${NC}"
    cp .env.example .env
    echo -e "${RED}请修改 .env 文件中的配置后重新运行此脚本${NC}"
    exit 1
fi

# 检查必要的环境变量
source .env
if [ "$JWT_SECRET" == "your-super-secret-jwt-key-change-this" ]; then
    echo -e "${RED}错误: 请修改 JWT_SECRET 环境变量${NC}"
    exit 1
fi

# 函数: 显示帮助
show_help() {
    echo "用法: ./deploy.sh [命令]"
    echo ""
    echo "命令:"
    echo "  start       启动所有服务"
    echo "  stop        停止所有服务"
    echo "  restart     重启所有服务"
    echo "  update      更新并重启服务"
    echo "  logs        查看日志"
    echo "  status      查看服务状态"
    echo "  backup      备份数据库"
    echo "  restore     恢复数据库"
    echo "  clean       清理未使用的资源"
    echo "  help        显示此帮助"
}

# 函数: 启动服务
start_services() {
    echo -e "${GREEN}正在启动服务...${NC}"
    docker-compose -f $COMPOSE_FILE up -d
    echo -e "${GREEN}服务启动完成${NC}"
    show_status
}

# 函数: 停止服务
stop_services() {
    echo -e "${YELLOW}正在停止服务...${NC}"
    docker-compose -f $COMPOSE_FILE down
    echo -e "${GREEN}服务已停止${NC}"
}

# 函数: 重启服务
restart_services() {
    echo -e "${YELLOW}正在重启服务...${NC}"
    docker-compose -f $COMPOSE_FILE restart
    echo -e "${GREEN}服务重启完成${NC}"
}

# 函数: 更新服务
update_services() {
    echo -e "${GREEN}正在更新服务...${NC}"
    
    # 拉取最新镜像
    docker-compose -f $COMPOSE_FILE pull
    
    # 重新构建并启动
    docker-compose -f $COMPOSE_FILE up -d --build
    
    # 清理旧镜像
    docker image prune -f
    
    echo -e "${GREEN}服务更新完成${NC}"
}

# 函数: 查看日志
show_logs() {
    docker-compose -f $COMPOSE_FILE logs -f --tail=100
}

# 函数: 查看状态
show_status() {
    echo -e "${GREEN}服务状态:${NC}"
    docker-compose -f $COMPOSE_FILE ps
    echo ""
    echo -e "${GREEN}资源使用:${NC}"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}"
}

# 函数: 备份数据库
backup_database() {
    BACKUP_DIR="./backups"
    BACKUP_FILE="$BACKUP_DIR/backup_$(date +%Y%m%d_%H%M%S).sql"
    
    mkdir -p $BACKUP_DIR
    
    echo -e "${GREEN}正在备份数据库...${NC}"
    docker-compose -f $COMPOSE_FILE exec -T postgres pg_dump -U $DB_USER $DB_NAME > $BACKUP_FILE
    
    # 压缩备份文件
    gzip $BACKUP_FILE
    
    echo -e "${GREEN}数据库已备份到: ${BACKUP_FILE}.gz${NC}"
    
    # 删除30天前的备份
    find $BACKUP_DIR -name "*.gz" -mtime +30 -delete
}

# 函数: 恢复数据库
restore_database() {
    if [ -z "$1" ]; then
        echo -e "${RED}请指定备份文件${NC}"
        echo "用法: ./deploy.sh restore <backup_file.sql.gz>"
        exit 1
    fi
    
    BACKUP_FILE=$1
    
    if [ ! -f "$BACKUP_FILE" ]; then
        echo -e "${RED}备份文件不存在: $BACKUP_FILE${NC}"
        exit 1
    fi
    
    echo -e "${YELLOW}警告: 这将覆盖当前数据库，是否继续? (y/n)${NC}"
    read -r confirm
    if [ "$confirm" != "y" ]; then
        echo "已取消"
        exit 0
    fi
    
    echo -e "${GREEN}正在恢复数据库...${NC}"
    gunzip -c $BACKUP_FILE | docker-compose -f $COMPOSE_FILE exec -T postgres psql -U $DB_USER $DB_NAME
    
    echo -e "${GREEN}数据库恢复完成${NC}"
}

# 函数: 清理资源
clean_resources() {
    echo -e "${YELLOW}正在清理未使用的 Docker 资源...${NC}"
    docker system prune -f
    docker volume prune -f
    echo -e "${GREEN}清理完成${NC}"
}

# 主逻辑
case "$1" in
    start)
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    update)
        update_services
        ;;
    logs)
        show_logs
        ;;
    status)
        show_status
        ;;
    backup)
        backup_database
        ;;
    restore)
        restore_database "$2"
        ;;
    clean)
        clean_resources
        ;;
    help|*)
        show_help
        ;;
esac
