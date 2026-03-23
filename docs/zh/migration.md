# 迁移指南

本指南介绍如何从 v1alpha1 迁移到 v1alpha2 API 版本。

## 概述

`v1alpha2` API 版本引入了多项改进：
- 使用标准 Kubernetes 条件增强状态报告
- 改进的资源管理和采用机制
- 新增 Kubernetes 集成 CRD（TunnelIngressClassConfig、TunnelGatewayClassConfig）
- 更好的错误处理和验证

## 自动转换

Operator 包含一个转换 webhook，可自动在 v1alpha1 和 v1alpha2 之间转换资源。这意味着：

- **现有 v1alpha1 资源** 无需修改即可继续工作
- **新资源** 应使用 v1alpha2
- **存储版本** 是 v1alpha2（资源以此格式存储）

## API 变更

### Tunnel / ClusterTunnel

无破坏性变更。以下字段保持不变：
- `spec.newTunnel`
- `spec.existingTunnel`
- `spec.cloudflare`
- `spec.size`
- `spec.image`

### TunnelBinding（已弃用）

`TunnelBinding` 仅为向后兼容保留。

- 旧版 API 组：`networking.cfargotunnel.com/v1alpha1`
- 兼容 API 组：`networking.cloudflare-operator.io/v1alpha2`
- 新部署推荐路径：使用 Ingress（`TunnelIngressClassConfig`）或 Gateway API（`TunnelGatewayClassConfig`）

### 状态条件

v1alpha2 使用标准 Kubernetes 条件类型：

| 条件 | 含义 |
|------|------|
| `Ready` | 资源完全可操作 |
| `Progressing` | 资源正在调和中 |
| `Degraded` | 资源有错误 |

## 迁移步骤

### 步骤 1：更新 Operator

确保运行 GitHub Releases 当前可用的 Operator 版本（例如当前可能是 `v0.0.1`）：

```bash
# 首先更新 CRDs
kubectl apply -f https://github.com/0ekk/cloudflare-operator/releases/latest/download/cloudflare-operator-crds.yaml

# 然后更新 operator
kubectl apply -f https://github.com/0ekk/cloudflare-operator/releases/latest/download/cloudflare-operator-no-webhook.yaml
```

### 步骤 2：验证转换 Webhook

检查转换 webhook 是否正在运行：

```bash
kubectl get pods -n cloudflare-operator-system
kubectl logs -n cloudflare-operator-system deployment/cloudflare-operator-controller-manager
```

### 步骤 3：测试现有资源

现有 v1alpha1 资源应继续工作：

```bash
kubectl get tunnels.networking.cloudflare-operator.io -A
kubectl get clustertunnels.networking.cloudflare-operator.io
```

### 步骤 4：迁移清单（可选）

更新清单以在新部署中使用 v1alpha2：

```yaml
# 之前（v1alpha1）
apiVersion: networking.cloudflare-operator.io/v1alpha1
kind: Tunnel

# 之后（v1alpha2）
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: Tunnel
```

### 步骤 5：迁移到 Ingress/Gateway API（推荐）

新部署建议使用 Ingress 或 Gateway API，不再推荐 TunnelBinding。

```yaml
# 推荐的 IngressClass parameters
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: cf-tunnel
spec:
  controller: cloudflare-operator.io/ingress-controller
  parameters:
    apiGroup: networking.cloudflare-operator.io
    kind: TunnelIngressClassConfig
    name: cf-tunnel
    scope: Namespace
    namespace: default
```

如果仍在使用 TunnelBinding，建议先从旧版 API 组迁移到 `networking.cloudflare-operator.io/v1alpha2`，再逐步迁移到 Ingress/Gateway API。

从 TunnelBinding 迁移到 Ingress/Gateway API 的完整步骤请参考 [TunnelBinding 迁移指南](migration/tunnelbinding-migration.md)。

## 回滚

如果遇到问题：

1. 转换 webhook 允许双向转换
2. 您可以继续使用 v1alpha1 资源
3. 检查 operator 日志以了解转换错误

## 故障排除

### 转换错误

如果资源转换失败：

```bash
# 检查 webhook 日志
kubectl logs -n cloudflare-operator-system deployment/cloudflare-operator-controller-manager | grep conversion

# 描述资源
kubectl describe tunnel <name> -n <namespace>
```

### 版本不匹配

如果看到版本不匹配错误：

1. 确保 CRD 已更新：`kubectl apply -f cloudflare-operator-crds.yaml`
2. 重启 operator：`kubectl rollout restart deployment -n cloudflare-operator-system`

## 常见问题

**问：需要重新创建资源吗？**
答：不需要，现有资源会自动转换。

**问：可以同时使用 v1alpha1 和 v1alpha2 吗？**
答：可以，转换 webhook 会自动处理。

**问：v1alpha1 什么时候会被移除？**
答：目前没有时间表。我们会在弃用前提前通知。
