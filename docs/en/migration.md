# Migration Guide

This guide covers migrating from v1alpha1 to v1alpha2 API version.

## Overview

The `v1alpha2` API version introduces several improvements:
- Enhanced status reporting with standard Kubernetes conditions
- Improved resource management and adoption
- New CRDs for Kubernetes integration (TunnelIngressClassConfig, TunnelGatewayClassConfig)
- Better error handling and validation

## Automatic Conversion

The operator includes a conversion webhook that automatically converts resources between v1alpha1 and v1alpha2. This means:

- **Existing v1alpha1 resources** continue to work without modification
- **New resources** should use v1alpha2
- **Storage version** is v1alpha2 (resources are stored in this format)

## API Changes

### Tunnel / ClusterTunnel

No breaking changes. The following fields are the same:
- `spec.newTunnel`
- `spec.existingTunnel`
- `spec.cloudflare`
- `spec.size`
- `spec.image`

### TunnelBinding (Deprecated)

`TunnelBinding` is kept for backward compatibility only.

- Legacy API group: `networking.cfargotunnel.com/v1alpha1`
- Compatibility API group: `networking.cloudflare-operator.io/v1alpha2`
- Recommended path for new setups: use Ingress (`TunnelIngressClassConfig`) or Gateway API (`TunnelGatewayClassConfig`)

### Status Conditions

v1alpha2 uses standard Kubernetes condition types:

| Condition | Meaning |
|-----------|---------|
| `Ready` | Resource is fully operational |
| `Progressing` | Resource is being reconciled |
| `Degraded` | Resource has errors |

## Migration Steps

### Step 1: Update Operator

Ensure you're running the currently available GitHub Release version (for example, this may be `v0.0.1`):

```bash
# Update CRDs first
kubectl apply -f https://github.com/0ekk/cloudflare-operator/releases/latest/download/cloudflare-operator-crds.yaml

# Then update operator
kubectl apply -f https://github.com/0ekk/cloudflare-operator/releases/latest/download/cloudflare-operator-no-webhook.yaml
```

### Step 2: Verify Conversion Webhook

Check that the conversion webhook is running:

```bash
kubectl get pods -n cloudflare-operator-system
kubectl logs -n cloudflare-operator-system deployment/cloudflare-operator-controller-manager
```

### Step 3: Test Existing Resources

Your existing v1alpha1 resources should continue to work:

```bash
kubectl get tunnels.networking.cloudflare-operator.io -A
kubectl get clustertunnels.networking.cloudflare-operator.io
```

### Step 4: Migrate Manifests (Optional)

Update your manifests to use v1alpha2 for new deployments:

```yaml
# Before (v1alpha1)
apiVersion: networking.cloudflare-operator.io/v1alpha1
kind: Tunnel

# After (v1alpha2)
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: Tunnel
```

### Step 5: Migrate to Ingress/Gateway API (Recommended)

For new deployments, use Ingress or Gateway API instead of TunnelBinding.

```yaml
# Recommended IngressClass parameters
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

If you still run TunnelBinding, migrate the legacy API group to `networking.cloudflare-operator.io/v1alpha2` first, then move to Ingress/Gateway API.

For detailed steps, see [TunnelBinding Migration Guide](migration/tunnelbinding-migration.md).

## Rollback

If you encounter issues:

1. The conversion webhook allows bidirectional conversion
2. You can continue using v1alpha1 resources
3. Check operator logs for conversion errors

## Troubleshooting

### Conversion Errors

If resources fail to convert:

```bash
# Check webhook logs
kubectl logs -n cloudflare-operator-system deployment/cloudflare-operator-controller-manager | grep conversion

# Describe the resource
kubectl describe tunnel <name> -n <namespace>
```

### Version Mismatch

If you see version mismatch errors:

1. Ensure CRDs are updated: `kubectl apply -f cloudflare-operator-crds.yaml`
2. Restart the operator: `kubectl rollout restart deployment -n cloudflare-operator-system`

## FAQ

**Q: Do I need to recreate my resources?**
A: No, existing resources are automatically converted.

**Q: Can I use both v1alpha1 and v1alpha2?**
A: Yes, the conversion webhook handles this automatically.

**Q: When will v1alpha1 be removed?**
A: No timeline yet. We'll provide advance notice before deprecation.
