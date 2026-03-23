# RedirectRule

RedirectRule 是一个命名空间级资源，用于在 Cloudflare 边缘创建 HTTP 重定向规则。

## 概述

RedirectRule 支持基于 URL 匹配模式把请求重定向到目标地址，重定向在 Cloudflare 边缘完成，无需回源。

### 主要特性

- URL 重定向
- 模式匹配
- HTTP 状态码控制
- 保留路径和参数

## 规范

| 字段 | 类型 | 必需 | 描述 |
|------|------|------|------|
| `pattern` | string | **是** | URL 匹配模式 |
| `destination` | string | **是** | 重定向目标 |
| `statusCode` | int | 否 | HTTP 状态码 |
| `cloudflare` | CloudflareDetails | **是** | Cloudflare API 凭证 |

## 示例

### 示例 1：HTTP 跳转到 HTTPS

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: RedirectRule
metadata:
  name: https-redirect
  namespace: production
spec:
  pattern: "http://example.com/*"
  destination: "https://example.com/$1"
  statusCode: 301
  cloudflare:
    credentialsRef:
      name: production
```

## 另请参阅

- [Cloudflare Redirect Rules](https://developers.cloudflare.com/rules/redirect/)
