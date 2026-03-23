# GatewayList

GatewayList 是一个集群级资源，用于定义可复用的 Gateway 列表。

## 概述

GatewayList 允许你创建域名或 IP 列表，并在多个 GatewayRule 中复用。

### 主要特性

- 可复用域名/IP 列表
- 集中式列表管理
- 多种列表类型
- 易于维护

## 规范

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `name` | string | **是** | 列表名称 |
| `type` | string | **是** | 列表类型：domains、ips |
| `items` | []string | **是** | 列表项 |
| `cloudflare` | CloudflareDetails | **是** | Cloudflare API 凭证 |

## 示例

### 示例 1：恶意域名列表

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: GatewayList
metadata:
  name: blocked-domains
spec:
  name: "Blocked Domains"
  type: "domains"
  items:
    - "malware.com"
    - "phishing.com"
  cloudflare:
    credentialsRef:
      name: production
```

## 相关资源

- [GatewayRule](gatewayrule.md) - 使用该列表的规则

## 另请参阅

- [Cloudflare Gateway Lists](https://developers.cloudflare.com/cloudflare-one/policies/gateway/)
