# DeviceSettingsPolicy

DeviceSettingsPolicy 是一个集群级资源，用于配置 Cloudflare WARP 客户端的设备策略。

## 概述

DeviceSettingsPolicy 允许你从 Kubernetes 集中配置 WARP 客户端行为和安全设置。

### 主要特性

- 设备设置管理
- 基于策略的配置
- 远程配置能力
- 合规控制

## 规范

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `name` | string | **是** | 策略名称 |
| `settings` | map[string]string | 否 | 设备设置项 |
| `cloudflare` | CloudflareDetails | **是** | Cloudflare API 凭证 |

## 示例

### 示例 1：设备设置策略

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: DeviceSettingsPolicy
metadata:
  name: enterprise-policy
spec:
  name: "Enterprise Policy"
  settings:
    splitTunneling: "enabled"
  cloudflare:
    credentialsRef:
      name: production
```

## 另请参阅

- [Cloudflare Device Settings](https://developers.cloudflare.com/warp-client/)
