# CI/CD 配置说明

本项目支持多种 CI/CD 平台，选择适合你的平台使用。

## 📁 配置文件

| 平台 | 配置文件 |
|------|----------|
| GitHub Actions | `.github/workflows/ci-cd.yaml` |
| GitLab CI | `.gitlab-ci.yml` |
| CircleCI | `.circleci/config.yml` |

## 🔄 流水线流程

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────────┐
│  Test   │───▶│  Build  │───▶│ Docker  │───▶│   Deploy    │
│  代码检查 │    │  构建   │    │  镜像   │    │  K8s部署    │
└─────────┘    └─────────┘    └─────────┘    └─────────────┘
```

## 🚀 触发条件

| 事件 | 动作 |
|------|------|
| Push 到 `develop` | 测试 → 构建 → 部署到**开发环境** |
| Push 到 `main` | 测试 → 构建 → Docker 镜像 |
| 创建 Tag `v*` | 测试 → 构建 → 部署到**生产环境** + Release |
| Pull Request | 仅运行测试 |

## ⚙️ 需要配置的 Secrets

### GitHub Actions

在 GitHub 仓库 Settings → Secrets and variables → Actions 中添加：

| Secret 名称 | 说明 |
|-------------|------|
| `KUBE_CONFIG_DEV` | 开发环境 kubeconfig (Base64 编码) |
| `KUBE_CONFIG_PROD` | 生产环境 kubeconfig (Base64 编码) |

### GitLab CI

在 GitLab 项目 Settings → CI/CD → Variables 中添加：

| Variable 名称 | 说明 |
|---------------|------|
| `KUBE_CONFIG_DEV` | 开发环境 kubeconfig (Base64 编码) |
| `KUBE_CONFIG_PROD` | 生产环境 kubeconfig (Base64 编码) |

### 生成 Base64 编码的 kubeconfig

```bash
# Linux/Mac
cat ~/.kube/config | base64 -w 0

# Windows PowerShell
[Convert]::ToBase64String([IO.File]::ReadAllBytes("$HOME\.kube\config"))
```

## 📝 使用示例

### 开发流程

```bash
# 1. 创建功能分支
git checkout -b feature/new-feature

# 2. 开发并提交
git add .
git commit -m "feat: add new feature"

# 3. 推送并创建 PR
git push origin feature/new-feature
# PR 会自动触发测试

# 4. 合并到 develop 分支
# 自动部署到开发环境

# 5. 合并到 main 分支
# 构建 Docker 镜像

# 6. 创建 Release Tag
git tag v1.0.0
git push origin v1.0.0
# 自动部署到生产环境
```

### 手动部署

```bash
# 部署到开发环境
kubectl apply -f deploy/k8s/ --context=dev-cluster

# 部署到生产环境
kubectl apply -f deploy/k8s/ --context=prod-cluster
```

## 🔧 自定义配置

### 修改 Docker Registry

编辑 `.github/workflows/ci-cd.yaml`：

```yaml
env:
  REGISTRY: your-registry.com  # 修改为你的镜像仓库
  IMAGE_NAME: your-org/go-zero-gorm-code
```

### 修改 K8s 命名空间

编辑 `deploy/k8s/namespace.yaml` 和其他 K8s 配置文件中的 namespace。

## 📊 监控和通知

可以添加以下集成：

- **Slack 通知**: 部署成功/失败时发送通知
- **钉钉通知**: 适合国内团队
- **邮件通知**: 发送部署报告

### 添加 Slack 通知示例 (GitHub Actions)

```yaml
- name: Notify Slack
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```
