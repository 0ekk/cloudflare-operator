# cloudflare-operator

Helm chart for deploying Cloudflare Zero Trust Operator.

This chart defaults to `config/default-no-webhook` behavior (operator deployment without webhook/cert-manager resources), and supports enabling webhook/cert-manager through values.

## Install

```bash
helm upgrade --install cloudflare-operator ./charts/cloudflare-operator \
  --namespace cloudflare-operator-system \
  --create-namespace
```

## Common values

```yaml
installCRDs: true

image:
  repository: ghcr.io/0ekk/cloudflare-operator
  tag: "v0.34.2"
  pullPolicy: IfNotPresent

operator:
  replicas: 1

webhook:
  enabled: false

certManager:
  enabled: false
```

## Upgrade

```bash
helm upgrade cloudflare-operator ./charts/cloudflare-operator \
  --namespace cloudflare-operator-system
```

Enable webhook + cert-manager:

```bash
helm upgrade --install cloudflare-operator ./charts/cloudflare-operator \
  --namespace cloudflare-operator-system \
  --set webhook.enabled=true \
  --set certManager.enabled=true
```

## Uninstall

```bash
helm uninstall cloudflare-operator -n cloudflare-operator-system
```
