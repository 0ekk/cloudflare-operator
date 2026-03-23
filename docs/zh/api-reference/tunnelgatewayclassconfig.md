# TunnelGatewayClassConfig

TunnelGatewayClassConfig 是一个命名空间作用域资源，用于配置 Kubernetes Gateway API 与 Cloudflare Tunnel 的集成。

## 概述

TunnelGatewayClassConfig 用于把 GatewayClass 关联到 Tunnel 或 ClusterTunnel，并配置 DNS 与回源默认行为。

### 主要特性

- Gateway API 集成
- 自动 DNS 管理
- 现代网络 API

## 规范

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `tunnelRef` | TunnelReference | **是** | 目标 `Tunnel` 或 `ClusterTunnel` |
| `defaultOriginRequest` | OriginRequestSpec | 否 | 后端连接的默认回源配置 |
| `dnsManagement` | string | 否 | DNS 管理模式：`Automatic`、`Manual`、`DNSRecord` |
| `dnsProxied` | bool | 否 | 管理的 DNS 记录是否走代理（默认 `true`） |
| `watchNamespaces` | []string | 否 | 限定监听的 Route 命名空间（空表示全部） |
| `fallbackTarget` | string | 否 | 未匹配请求的兜底目标（默认 `http_status:404`） |

## 示例

### 示例 1：TunnelGatewayClassConfig

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: TunnelGatewayClassConfig
metadata:
  name: cf-gateway
  namespace: default
spec:
  tunnelRef:
    kind: ClusterTunnel
    name: my-cluster-tunnel
  dnsManagement: Automatic
  dnsProxied: true
  fallbackTarget: http_status:404
```

### 示例 2：GatewayClass 参数

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: cf-gateway
spec:
  controllerName: cloudflare-operator.io/gateway-controller
  parametersRef:
    group: networking.cloudflare-operator.io
    kind: TunnelGatewayClassConfig
    name: cf-gateway
    namespace: default
```

`TunnelGatewayClassConfig` 是命名空间作用域资源，因此 GatewayClass 的 `parametersRef.namespace` 需要指向该配置所在命名空间。

## 另请参阅

- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
