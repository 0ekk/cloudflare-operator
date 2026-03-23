# API Reference

This section contains detailed documentation for all Custom Resource Definitions (CRDs).

## CRD Categories

### Core Configuration
- [CloudflareCredentials](cloudflarecredentials.md) - Shared Cloudflare API credentials
- [CloudflareDomain](cloudflaredomain.md) - Zone/domain-level configuration

### Tunnel Management
- [Tunnel](tunnel.md) - Namespace-scoped Cloudflare Tunnel
- [ClusterTunnel](clustertunnel.md) - Cluster-wide Cloudflare Tunnel
- [TunnelBinding](tunnelbinding.md) - Deprecated legacy Service-to-Tunnel binding

### Private Network
- [VirtualNetwork](virtualnetwork.md) - Traffic isolation network
- [NetworkRoute](networkroute.md) - CIDR routing through tunnel
- [PrivateService](privateservice.md) - Private IP service exposure
- [WARPConnector](warpconnector.md) - Site-to-site WARP connector

### Access Control
- [AccessApplication](accessapplication.md) - Zero Trust application
- [AccessGroup](accessgroup.md) - Reusable access policy group
- [AccessPolicy](accesspolicy.md) - Reusable access policy template
- [AccessIdentityProvider](accessidentityprovider.md) - Identity provider config
- [AccessServiceToken](accessservicetoken.md) - M2M authentication token

### Gateway & Security
- [GatewayRule](gatewayrule.md) - DNS/HTTP/L4 policy rule
- [GatewayList](gatewaylist.md) - List for gateway rules
- [GatewayConfiguration](gatewayconfiguration.md) - Global gateway settings
- [ZoneRuleset](zoneruleset.md) - Zone ruleset management
- [TransformRule](transformrule.md) - Edge request/response transforms
- [RedirectRule](redirectrule.md) - Edge URL redirects

### Device Management
- [DeviceSettingsPolicy](devicesettingspolicy.md) - WARP client configuration
- [DevicePostureRule](deviceposturerule.md) - Device health check rule

### DNS & Connectivity
- [DNSRecord](dnsrecord.md) - DNS record management
- [OriginCACertificate](origincacertificate.md) - Origin CA certificate management

### Storage
- [R2Bucket](r2bucket.md) - R2 bucket management
- [R2BucketDomain](r2bucketdomain.md) - Custom domain for R2 bucket
- [R2BucketNotification](r2bucketnotification.md) - R2 event notifications

### Pages & Workers
- [PagesProject](pagesproject.md) - Cloudflare Pages project management
- [PagesDeployment](pagesdeployment.md) - Deploy versions to Pages
- [PagesDomain](pagesdomain.md) - Custom domain for Pages

### Registrar
- [DomainRegistration](domainregistration.md) - Domain registration (Enterprise)

### Kubernetes Integration
- [TunnelIngressClassConfig](tunnelingressclassconfig.md) - Ingress integration
- [TunnelGatewayClassConfig](tunnelgatewayclassconfig.md) - Gateway API integration

## Common Types

### CloudflareSpec

All CRDs that interact with Cloudflare API include a `cloudflare` spec:

```yaml
spec:
  cloudflare:
    credentialsRef:
      name: default
```

### Status Conditions

All CRDs report status through standard Kubernetes conditions:

| Condition | Description |
|-----------|-------------|
| `Ready` | Resource is fully reconciled and operational |
| `Progressing` | Resource is being created or updated |
| `Degraded` | Resource has errors but may be partially functional |

## API Version

Current API version: `networking.cloudflare-operator.io/v1alpha2`

Legacy version `v1alpha1` is deprecated but still supported for backwards compatibility.
