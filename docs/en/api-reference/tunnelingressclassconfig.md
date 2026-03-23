# TunnelIngressClassConfig

TunnelIngressClassConfig is a namespaced resource that configures Kubernetes Ingress integration with Cloudflare Tunnels.

## Overview

TunnelIngressClassConfig links an IngressClass to a Tunnel or ClusterTunnel and controls DNS/origin defaults for Ingress routes.

### Key Features

- Ingress integration
- Automatic DNS management
- Origin request defaults

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tunnelRef` | TunnelReference | **Yes** | Target `Tunnel` or `ClusterTunnel` |
| `defaultProtocol` | string | No | Default backend protocol (`http` by default) |
| `defaultOriginRequest` | OriginRequestSpec | No | Default origin settings for backend connections |
| `dnsManagement` | string | No | DNS mode: `Automatic`, `Manual`, `DNSRecord` |
| `dnsProxied` | bool | No | Whether managed DNS records are proxied (default: `true`) |
| `watchNamespaces` | []string | No | Limit watched Ingress namespaces (empty = all) |

## Examples

### Example 1: TunnelIngressClassConfig

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

### Example 2: IngressClass Parameters

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

`TunnelIngressClassConfig` is namespaced, so IngressClass parameters should normally use `scope: Namespace` and set `namespace`.

## See Also

- [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)
