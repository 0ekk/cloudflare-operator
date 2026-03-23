# cloudflare-operator

Helm chart for deploying Cloudflare Zero Trust Operator.

This chart defaults to `config/default-no-webhook` behavior (operator deployment without webhook/cert-manager resources), and supports enabling webhook/cert-manager through values.

## Install

```bash
helm repo add cloudflare-operator https://0ekk.github.io/cloudflare-operator
helm repo update

helm upgrade --install cloudflare-operator cloudflare-operator/cloudflare-operator \
  --namespace cloudflare-operator-system \
  --create-namespace
```

## Common values

```yaml
installCRDs: true

image:
  repository: ghcr.io/0ekk/cloudflare-operator
  # Empty means default to <Chart.Version>
  tag: "0.34.2"
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
helm upgrade cloudflare-operator cloudflare-operator/cloudflare-operator \
  --namespace cloudflare-operator-system
```

Enable webhook + cert-manager:

```bash
helm upgrade --install cloudflare-operator cloudflare-operator/cloudflare-operator \
  --namespace cloudflare-operator-system \
  --set webhook.enabled=true \
  --set certManager.enabled=true
```

## Uninstall

```bash
helm uninstall cloudflare-operator -n cloudflare-operator-system
```
