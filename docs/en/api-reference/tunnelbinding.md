# TunnelBinding

DEPRECATED: TunnelBinding is a namespaced resource. Please use Ingress or Gateway API instead.

## Overview

TunnelBinding is deprecated. It was used to bind Tunnels to services. Please migrate to standard Kubernetes Ingress or Gateway API resources.
TunnelBinding should be treated as legacy compatibility only and is not recommended for new production deployments.

### Alternatives

- Use Kubernetes Ingress with TunnelIngressClassConfig
- Use Kubernetes Gateway API with TunnelGatewayClassConfig
- Use DNSRecord resources for manual DNS management

## See Also

- [TunnelBinding Migration Guide](../migration/tunnelbinding-migration.md)
- [Kubernetes Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
