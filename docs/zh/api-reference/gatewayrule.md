# GatewayRule

GatewayRule 是一个集群级资源，用于定义 Cloudflare Gateway 的 DNS 过滤与安全策略规则。

## 概述

GatewayRule 用于在 Cloudflare Gateway 中创建 DNS 过滤和安全策略。你可以基于域名模式等条件对查询执行 block、allow、redirect 等动作。

### 主要特性

- DNS 过滤规则
- 安全策略执行
- 流量路由控制
- 基于模式匹配
- 多种动作类型

## 规范

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `pattern` | string | **是** | 匹配的域名模式 |
| `action` | string | **是** | 动作：block、allow、redirect |
| `priority` | int | 否 | 规则优先级 |
| `cloudflare` | CloudflareDetails | **是** | Cloudflare API 凭证 |

## 示例

### 示例 1：阻止恶意域名

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: GatewayRule
metadata:
  name: block-malware
spec:
  pattern: "*.malware-domain.com"
  action: "block"
  priority: 10
  cloudflare:
    credentialsRef:
      name: production
```

## 相关资源

- [GatewayList](gatewaylist.md) - 规则列表
- [GatewayConfiguration](gatewayconfiguration.md) - Gateway 全局设置

## 另请参阅

- [Cloudflare Gateway DNS](https://developers.cloudflare.com/cloudflare-one/policies/gateway/dns-policies/)
