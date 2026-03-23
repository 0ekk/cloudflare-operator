# TunnelGatewayClassConfig

TunnelGatewayClassConfig is a namespaced resource that configures Kubernetes Gateway API integration with Cloudflare Tunnels.

## Overview

TunnelGatewayClassConfig links a GatewayClass to a Tunnel or ClusterTunnel and controls DNS/origin defaults for Gateway API routes.

### Key Features

- Gateway API integration
- Automatic DNS management
- Modern networking APIs

## Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tunnelRef` | TunnelReference | **Yes** | Target `Tunnel` or `ClusterTunnel` |
| `defaultOriginRequest` | OriginRequestSpec | No | Default origin settings for backend connections |
| `dnsManagement` | string | No | DNS mode: `Automatic`, `Manual`, `DNSRecord` |
| `dnsProxied` | bool | No | Whether managed DNS records are proxied (default: `true`) |
| `watchNamespaces` | []string | No | Limit watched Route namespaces (empty = all) |
| `fallbackTarget` | string | No | Fallback target for unmatched requests (default `http_status:404`) |

## Examples

### Example 1: TunnelGatewayClassConfig

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: TunnelGatewayClassConfig
metadata:
  name: cf-gateway
  namespace: default
spec:
  tunnelRef:
    kind: ClusterTunnel
    name: my-cluster-tunnel
  dnsManagement: Automatic
  dnsProxied: true
  fallbackTarget: http_status:404
```

### Example 2: GatewayClass Parameters

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: cf-gateway
spec:
  controllerName: cloudflare-operator.io/gateway-controller
  parametersRef:
    group: networking.cloudflare-operator.io
    kind: TunnelGatewayClassConfig
    name: cf-gateway
    namespace: default
```

`TunnelGatewayClassConfig` is namespaced, so GatewayClass `parametersRef.namespace` should point to the namespace where the config exists.

## See Also

- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
