# Getting Started

This guide will help you install the Cloudflare Operator and create your first tunnel.

## Prerequisites

- Kubernetes cluster v1.28+
- `kubectl` configured with cluster access
- Cloudflare account with Zero Trust enabled
- Cloudflare API Token

## Installation

Choose one of the following installation methods:

### Option A: Full Installation (Recommended for new users)

Single command to install everything:

```bash
# All-in-one: CRDs + Namespace + RBAC + Operator
kubectl apply -f https://github.com/0ekk/cloudflare-operator/releases/latest/download/cloudflare-operator-full-no-webhook.yaml
```

### Option B: Modular Installation (Recommended for production)

For fine-grained control over installation:

```bash
# Step 1: Install CRDs (requires cluster-admin)
kubectl apply -f https://github.com/0ekk/cloudflare-operator/releases/latest/download/cloudflare-operator-crds.yaml

# Step 2: Create namespace
kubectl apply -f https://github.com/0ekk/cloudflare-operator/releases/latest/download/cloudflare-operator-namespace.yaml

# Step 3: Install operator (RBAC + Deployment)
kubectl apply -f https://github.com/0ekk/cloudflare-operator/releases/latest/download/cloudflare-operator-no-webhook.yaml
```

### Available Installation Files

| File | Contents | Use Case |
|------|----------|----------|
| `cloudflare-operator-full.yaml` | CRDs + Namespace + RBAC + Operator + Webhook | Full with cert-manager |
| `cloudflare-operator-full-no-webhook.yaml` | CRDs + Namespace + RBAC + Operator | Full without webhook |
| `cloudflare-operator-crds.yaml` | CRDs only | Modular installation |
| `cloudflare-operator-namespace.yaml` | Namespace only | Modular installation |
| `cloudflare-operator.yaml` | RBAC + Operator + Webhook | Upgrade existing installation |
| `cloudflare-operator-no-webhook.yaml` | RBAC + Operator | Upgrade without webhook |

### Verify Installation

```bash
# Check operator pod
kubectl get pods -n cloudflare-operator-system

# Check CRDs
kubectl get crds | grep cloudflare
```

Expected output:
```
NAME                                                              CREATED AT
accessapplications.networking.cloudflare-operator.io              2024-01-01T00:00:00Z
accessgroups.networking.cloudflare-operator.io                    2024-01-01T00:00:00Z
...
tunnels.networking.cloudflare-operator.io                         2024-01-01T00:00:00Z
```

## Create Your First Tunnel

### Step 1: Create API Credentials

1. Go to [Cloudflare Dashboard > API Tokens](https://dash.cloudflare.com/profile/api-tokens)
2. Create a Custom Token with these permissions:
   - `Account:Cloudflare Tunnel:Edit`
   - `Zone:DNS:Edit` (for your domain)

3. Create the Kubernetes Secret and CloudflareCredentials:

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloudflare-credentials
  namespace: cloudflare-operator-system
type: Opaque
stringData:
  CLOUDFLARE_API_TOKEN: "<your-api-token>"
---
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: CloudflareCredentials
metadata:
  name: default
spec:
  accountId: "<your-account-id>"
  authType: apiToken
  secretRef:
    name: cloudflare-credentials
    namespace: cloudflare-operator-system
  isDefault: true
```

```bash
kubectl apply -f secret.yaml
```

### Step 2: Find Your Account ID

1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Select any domain
3. Find **Account ID** in the right sidebar under "API"

### Step 3: Create a Tunnel

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: Tunnel
metadata:
  name: my-first-tunnel
  namespace: default
spec:
  newTunnel:
    name: my-k8s-tunnel
  cloudflare:
    domain: example.com
    credentialsRef:
      name: default
```

```bash
kubectl apply -f tunnel.yaml
```

### Step 4: Verify Tunnel

```bash
# Check tunnel status
kubectl get tunnel my-first-tunnel

# Check cloudflared deployment
kubectl get deployment -l app.kubernetes.io/name=cloudflared

# Check cloudflared logs
kubectl logs -l app.kubernetes.io/name=cloudflared
```

### Step 5: Expose a Service (Recommended: Ingress)

Deploy a sample application:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
spec:
  replicas: 1
  selector:
    matchLabels:
      app: hello-world
  template:
    metadata:
      labels:
        app: hello-world
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
          ports:
            - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: hello-world
spec:
  selector:
    app: hello-world
  ports:
    - port: 80
```

Create `TunnelIngressClassConfig` and `IngressClass`:

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
---
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
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hello-world
  namespace: default
spec:
  ingressClassName: cf-tunnel
  rules:
    - host: hello.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: hello-world
                port:
                  number: 80
```

```bash
kubectl apply -f ingress.yaml
```

`TunnelBinding` is deprecated and kept for backward compatibility only. For new setups, prefer Ingress or Gateway API.

### Step 6: Access Your Application

After a few moments, your application will be accessible at `https://hello.example.com`.

```bash
# Verify DNS record
dig hello.example.com

# Access the application
curl https://hello.example.com
```

## Advanced Configuration

### Scaling Tunnel Replicas

Use the `deployPatch` field to customize the cloudflared deployment. This is a JSON patch applied to the deployment spec.

**Set replica count:**

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: Tunnel
metadata:
  name: my-tunnel
  namespace: default
spec:
  newTunnel:
    name: my-k8s-tunnel
  cloudflare:
    domain: example.com
    credentialsRef:
      name: default
  deployPatch: '{"spec":{"replicas":3}}'
```

**Set resources and node selector:**

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: Tunnel
metadata:
  name: my-tunnel
  namespace: default
spec:
  newTunnel:
    name: my-k8s-tunnel
  cloudflare:
    domain: example.com
    credentialsRef:
      name: default
  deployPatch: |
    {
      "spec": {
        "replicas": 2,
        "template": {
          "spec": {
            "nodeSelector": {
              "node-role.kubernetes.io/edge": "true"
            },
            "containers": [{
              "name": "cloudflared",
              "resources": {
                "requests": {"cpu": "100m", "memory": "128Mi"},
                "limits": {"cpu": "500m", "memory": "512Mi"}
              }
            }]
          }
        }
      }
    }
```

### Using ClusterTunnel

For cluster-wide tunnels (accessible from any namespace), use ClusterTunnel:

```yaml
apiVersion: networking.cloudflare-operator.io/v1alpha2
kind: ClusterTunnel
metadata:
  name: shared-tunnel
spec:
  newTunnel:
    name: shared-k8s-tunnel
  cloudflare:
    domain: example.com
    credentialsRef:
      name: default
  deployPatch: '{"spec":{"replicas":2}}'
```

> **Note:** For ClusterTunnel and other cluster-scoped resources, use a cluster-scoped `CloudflareCredentials`. Its referenced Secret should be in `cloudflare-operator-system`.

## What's Next?

- [Configure API Token Permissions](configuration.md)
- [Enable Private Network Access](api-reference/networkroute.md)
- [Add Zero Trust Authentication](api-reference/accessapplication.md)
- [View All Examples](../../examples/)
