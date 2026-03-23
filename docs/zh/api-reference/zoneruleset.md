# ZoneRuleset

ZoneRuleset 是一个命名空间级资源，用于管理 Cloudflare 规则集（如 WAF、限速等安全能力）。

## 概述

ZoneRuleset 允许你在 zone 级别配置 Cloudflare 托管规则集，包括 WAF 规则、限速策略以及其他安全功能。

### 主要特性

- WAF 规则管理
- 限速配置
- 安全规则控制
- 规则版本管理

## 规范

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `name` | string | **是** | 规则集名称 |
| `kind` | string | **是** | 规则集类型（waf、rateLimit 等） |
| `rules` | []Rule | 否 | 要应用的规则 |
| `cloudflare` | CloudflareDetails | **是** | Cloudflare API 凭证 |

## 示例

### 示例 1：WAF 规则集

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: ZoneRuleset
metadata:
  name: waf-rules
  namespace: production
spec:
  name: "WAF Rules"
  kind: "waf"
  cloudflare:
    credentialsRef:
      name: production
```

## 另请参阅

- [Cloudflare Rulesets](https://developers.cloudflare.com/ruleset-engine/)
