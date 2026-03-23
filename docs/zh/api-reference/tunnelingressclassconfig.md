# TunnelIngressClassConfig

TunnelIngressClassConfig 是一个命名空间作用域资源，用于配置 Kubernetes Ingress 与 Cloudflare Tunnel 的集成。

## 概述

TunnelIngressClassConfig 用于把 IngressClass 关联到 Tunnel 或 ClusterTunnel，并配置 DNS 与回源默认行为。

### 主要特性

- Ingress 集成
- 自动 DNS 管理
- 回源请求默认配置

## 规范

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `tunnelRef` | TunnelReference | **是** | 目标 `Tunnel` 或 `ClusterTunnel` |
| `defaultProtocol` | string | 否 | 默认后端协议（默认 `http`） |
| `defaultOriginRequest` | OriginRequestSpec | 否 | 后端连接的默认回源配置 |
| `dnsManagement` | string | 否 | DNS 管理模式：`Automatic`、`Manual`、`DNSRecord` |
| `dnsProxied` | bool | 否 | 管理的 DNS 记录是否走代理（默认 `true`） |
| `watchNamespaces` | []string | 否 | 限定监听的 Ingress 命名空间（空表示全部） |

## 示例

### 示例 1：TunnelIngressClassConfig

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: TunnelIngressClassConfig
metadata:
  name: cf-tunnel
  namespace: default
spec:
  tunnelRef:
    kind: Tunnel
    name: my-first-tunnel
  dnsManagement: Automatic
  dnsProxied: true
```

### 示例 2：IngressClass 参数

```yaml
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

`TunnelIngressClassConfig` 是命名空间作用域资源，因此 IngressClass parameters 通常应使用 `scope: Namespace` 并设置 `namespace`。

## 另请参阅

- [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)
