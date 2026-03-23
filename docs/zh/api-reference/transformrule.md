# TransformRule

TransformRule 是一个命名空间级资源，用于在 Cloudflare 边缘修改 HTTP 请求和响应。

## 概述

TransformRule 允许你在请求到达源站前，在 Cloudflare 边缘对请求头、响应头及 URL 相关内容进行改写。

### 主要特性

- Header 改写
- URL 重写
- 请求/响应处理
- 边缘侧规则执行

## 规范

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `name` | string | **是** | 规则名称 |
| `pattern` | string | **是** | URL 匹配模式 |
| `transforms` | []Transform | 否 | 要应用的转换规则 |
| `cloudflare` | CloudflareDetails | **是** | Cloudflare API 凭证 |

## 示例

### 示例 1：添加安全响应头

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: TransformRule
metadata:
  name: add-security-headers
  namespace: production
spec:
  name: "Add Security Headers"
  pattern: "*example.com/*"
  cloudflare:
    credentialsRef:
      name: production
```

## 另请参阅

- [Cloudflare Transform Rules](https://developers.cloudflare.com/rules/transform/)
