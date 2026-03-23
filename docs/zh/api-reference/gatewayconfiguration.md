# GatewayConfiguration

GatewayConfiguration 是一个集群级资源，用于配置全局 Cloudflare Gateway 设置。

## 概述

GatewayConfiguration 管理账号级 Gateway 配置，包括日志、证书检查和全局策略。

### 主要特性

- 全局 Gateway 设置
- 日志配置
- 策略默认值
- 证书相关选项

## 规范

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `logging` | LoggingConfig | 否 | 日志设置 |
| `inspection` | bool | 否 | 证书检查 |
| `cloudflare` | CloudflareDetails | **是** | Cloudflare API 凭证 |

## 示例

### 示例 1：开启 Gateway 日志

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: GatewayConfiguration
metadata:
  name: gateway-config
spec:
  logging:
    enabled: true
    level: "standard"
  inspection: true
  cloudflare:
    credentialsRef:
      name: production
```

## 另请参阅

- [Cloudflare Gateway Configuration](https://developers.cloudflare.com/cloudflare-one/policies/gateway/)
