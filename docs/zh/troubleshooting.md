# 故障排除

本指南帮助诊断和解决 Cloudflare Operator 的常见问题。

## 诊断命令

### 检查 Operator 状态

```bash
# Operator pod 状态
kubectl get pods -n cloudflare-operator-system

# Operator 日志
kubectl logs -n cloudflare-operator-system deployment/cloudflare-operator-controller-manager

# 启用调试日志
kubectl patch deployment cloudflare-operator-controller-manager \
  -n cloudflare-operator-system \
  --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--zap-log-level=debug"}]'
```

### 检查资源状态

```bash
# 列出所有 operator 资源
kubectl get tunnels,clustertunnels,tunnelbindings,virtualnetworks,networkroutes -A

# 带条件的详细状态
kubectl get tunnel <name> -o jsonpath='{.status.conditions}' | jq

# 资源事件
kubectl describe tunnel <name>
```

## 常见问题

### Ingress 已应用但未生效

**症状：**
- Ingress 已存在，但 Cloudflare Tunnel 中没有发布对应路由
- Operator 持续 Reconcile，但没有创建/更新预期路由

**诊断步骤：**

```bash
# 检查 Ingress 和 IngressClass
kubectl get ingress <name> -n <namespace> -o yaml
kubectl get ingressclass <name> -o yaml

# 检查 TunnelIngressClassConfig
kubectl get tunnelingressclassconfig <name> -o yaml

# 过滤 operator 日志
kubectl logs -n cloudflare-operator-system deployment/cloudflare-operator-controller-manager | \
  grep -E "IngressClass|TunnelIngressClassConfig|resolve zone|CNAME|1056"
```

**常见原因：**

1. **IngressClass 参数无效**
   - 报错：`spec.parameters.namespace: Forbidden: parameters.scope is set to 'Cluster'`
   - 处理：
     - 如果使用 `scope: Cluster`，删除 `parameters.namespace`
     - 对于 `TunnelIngressClassConfig`（命名空间作用域），使用 `scope: Namespace` 并设置 `parameters.namespace`

2. **Zone / 凭证解析失败**
   - 报错：`Failed to resolve zone and credentials`
   - 处理：
     - 确认 `CloudflareCredentials` 存在且可读
     - 确认 `accountId` 正确
     - 确认 API token 包含 `Account:Cloudflare Tunnel:Edit` 和 zone 级 DNS 权限
     - 确认引用域名在同一账号下且为活动状态

3. **CNAME 内容非法**
   - 报错：`Content for CNAME record is invalid. (9007)`
   - 处理：
     - 确保 hostname 是有效 FQDN（不含 scheme/path，字段位置正确）
     - 确保生成的目标是合法 tunnel hostname
     - 确认 hostname 属于配置的 Cloudflare zone

4. **隧道配置因 ingress 规则为空被拒绝**
   - 报错：`The config file doesn't contain any ingress rules (1056)`
   - 该问题通常出现在删除所有应用路由后。
   - 处理：升级到会在同步隧道配置时始终保留兜底规则（`http_status:404`）的版本。

### 隧道无法连接

**症状：**
- 隧道状态显示错误
- cloudflared pods 未运行或 CrashLooping

**诊断步骤：**

```bash
# 检查隧道状态
kubectl get tunnel <name> -o wide

# 检查 cloudflared 部署
kubectl get deployment -l app.kubernetes.io/name=cloudflared

# 检查 cloudflared 日志
kubectl logs -l app.kubernetes.io/name=cloudflared
```

**常见原因：**

1. **API Token 无效**
   ```bash
   # 验证 token 是否有效
   curl -X GET "https://api.cloudflare.com/client/v4/user/tokens/verify" \
     -H "Authorization: Bearer $(kubectl get secret cloudflare-credentials -o jsonpath='{.data.CLOUDFLARE_API_TOKEN}' | base64 -d)"
   ```

2. **Account ID 错误**
   - 在 Cloudflare 控制台验证 account ID
   - 检查隧道规格中的拼写

3. **网络连接问题**
   - 确保 pods 能访问 `api.cloudflare.com` 和 `*.cloudflareaccess.com`
   - 检查网络策略

4. **Secret 未找到**
   ```bash
   kubectl get secret <secret-name> -n <namespace>
   ```

### DNS 记录未创建

**症状：**
- Ingress 已存在但 DNS 无法解析
- Cloudflare 控制台中没有 CNAME 记录

**诊断步骤：**

```bash
# 检查 Ingress 和 IngressClass
kubectl get ingress <name> -n <namespace> -o yaml
kubectl get ingressclass <name> -o yaml

# 检查 TunnelIngressClassConfig
kubectl get tunnelingressclassconfig <name> -n <namespace> -o yaml

# 检查 operator 日志中的 DNS 错误
kubectl logs -n cloudflare-operator-system deployment/cloudflare-operator-controller-manager | grep -i dns
```

**常见原因：**

1. **缺少 DNS:Edit 权限**
   - 检查 token 是否具有该区域的 `Zone:DNS:Edit` 权限

2. **域名错误**
   - 验证 `cloudflare.domain` 是否与你的 Cloudflare 区域匹配

3. **区域未找到**
   - 域名必须在你的 Cloudflare 账户中处于活动状态

4. **Legacy TunnelBinding 路径**
   - 如果仍在使用 TunnelBinding，请单独检查：
   ```bash
   kubectl describe tunnelbinding <name>
   ```

### 网络路由不工作

**症状：**
- WARP 客户端无法访问路由的 IP
- 流量未通过隧道

**诊断步骤：**

```bash
# 检查 NetworkRoute 状态
kubectl get networkroute <name> -o wide

# 验证隧道是否启用了 WARP 路由
kubectl get clustertunnel <name> -o jsonpath='{.spec.enableWarpRouting}'
```

**常见原因：**

1. **WARP 路由未启用**
   ```yaml
   spec:
     enableWarpRouting: true  # 必须为 true
   ```

2. **路由冲突**
   - 在 Cloudflare 控制台检查是否有重叠路由
   - 验证 CIDR 不与现有路由冲突

3. **虚拟网络不匹配**
   - WARP 客户端必须连接到正确的虚拟网络

### Access 应用未保护

**症状：**
- 应用无需认证即可访问
- 登录页面未显示

**诊断步骤：**

```bash
# 检查 AccessApplication 状态
kubectl get accessapplication <name> -o wide
kubectl describe accessapplication <name>
```

**常见原因：**

1. **DNS 未指向隧道**
   - 应用域名必须通过 Cloudflare Tunnel 提供服务

2. **策略配置错误**
   - 检查 AccessGroup 规则是否正确
   - 验证策略决策是 `allow` 而不是 `bypass`

3. **缺少 IdP 配置**
   - 必须配置并引用 AccessIdentityProvider

### 资源卡在删除中

**症状：**
- 资源设置了 `DeletionTimestamp`
- Finalizers 阻止删除

**诊断步骤：**

```bash
# 检查 finalizers
kubectl get <resource> <name> -o jsonpath='{.metadata.finalizers}'

# 检查删除错误
kubectl describe <resource> <name>
```

**解决方案：**

1. **检查 Cloudflare API 错误**
   - 资源可能已在 Cloudflare 中删除
   - Operator 会重试删除

2. **手动删除 Finalizer**（谨慎使用）
   ```bash
   kubectl patch <resource> <name> -p '{"metadata":{"finalizers":null}}' --type=merge
   ```

## 错误消息

### "API Token validation failed"

- Token 无效或已过期
- 在 Cloudflare 控制台重新创建 token

### "Zone not found"

- 域名不在你的 Cloudflare 账户中
- 域名未激活（等待名称服务器更改）

### "Conflict: resource already exists"

- Cloudflare 中已存在同名资源
- 使用 `existingTunnel` 采用现有资源

### "Permission denied"

- Token 缺少所需权限
- 检查[权限矩阵](configuration.md#权限矩阵)

### "Failed to resolve zone and credentials"

- 凭证引用错误、凭证不存在或命名空间不正确
- 域名与 token 所在账号中的活动 Cloudflare zone 不匹配
- token 缺少必需的账号级或 zone 级权限

### "Content for CNAME record is invalid. (9007)"

- DNS 记录目标不是有效的 CNAME 目标
- Ingress/Tunnel 配置中的 hostname 与 zone 不匹配

## 获取帮助

如果问题持续存在：

1. **收集日志**
   ```bash
   kubectl logs -n cloudflare-operator-system deployment/cloudflare-operator-controller-manager > operator.log
   ```

2. **检查 GitHub Issues**
   - 搜索[现有问题](https://github.com/0ekk/cloudflare-operator/issues)

3. **开新 Issue**
   - 包含 operator 版本
   - 包含相关 CRD 清单（已脱敏）
   - 包含错误日志
