# DevicePostureRule

DevicePostureRule 是一个集群级资源，用于定义 Zero Trust Access 的设备安全要求。

## 概述

DevicePostureRule 用于定义设备在访问受保护应用前必须满足的条件，例如防火墙状态、杀毒软件状态或磁盘加密状态。

### 主要特性

- 设备安全要求定义
- 多种条件类型
- 姿态检查与验证
- 自动策略执行

## 规范

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `name` | string | **是** | 规则名称 |
| `rules` | []PostureCheck | **是** | 姿态检查规则 |
| `cloudflare` | CloudflareDetails | **是** | Cloudflare API 凭证 |

## 示例

### 示例 1：要求防火墙开启

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: DevicePostureRule
metadata:
  name: require-antivirus
spec:
  name: "Require Antivirus"
  rules:
    - type: "firewall"
      enabled: true
  cloudflare:
    credentialsRef:
      name: production
```

## 另请参阅

- [Cloudflare Device Posture](https://developers.cloudflare.com/cloudflare-one/identity/devices/)
