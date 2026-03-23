# TunnelBinding

已弃用：TunnelBinding 是一个命名空间级资源。请改用 Ingress 或 Gateway API。

## 概述

TunnelBinding 已弃用。它曾用于将 Tunnel 绑定到服务。请迁移到标准 Kubernetes Ingress 或 Gateway API 资源。
TunnelBinding 应仅作为存量兼容方案使用，不建议用于新的生产部署。

### 替代方案

- 使用带有 TunnelIngressClassConfig 的 Kubernetes Ingress
- 使用带有 TunnelGatewayClassConfig 的 Kubernetes Gateway API
- 使用 DNSRecord 资源进行手动 DNS 管理

## 另请参阅

- [TunnelBinding 迁移指南](../migration/tunnelbinding-migration.md)
- [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
