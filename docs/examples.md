# Examples

This page provides examples of common Capsailer usage scenarios.

## Basic Bundle Creation

This example shows how to create a basic bundle with a few container images and Helm charts.

### Manifest File

```yaml
# manifest.yaml
images:
  - nginx:1.25
  - redis:7.0

charts:
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
```

### Build Command

```bash
capsailer init --manifest manifest.yaml
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz
```

## Deploying a Registry with Persistence

This example shows how to deploy a registry with persistent storage.

```bash
capsailer registry --namespace my-registry --storage-class standard --storage-size 20Gi
```

## Pushing Images to an External Registry

This example shows how to push images from a bundle to an external registry.

```bash
capsailer push --bundle capsailer-bundle.tar.gz --external-registry registry.example.com --username myuser --password mypassword
```

## Complete Air-Gapped Workflow

This example shows the complete workflow for air-gapped deployments.

### In the Connected Environment

```bash
# Create a manifest file
cat > manifest.yaml << EOF
images:
  - nginx:1.25
  - redis:7.0
  - bitnami/postgresql:15.4.0
  - registry:2

charts:
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
EOF

# Validate the manifest
capsailer init --manifest manifest.yaml

# Build the bundle
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz
```

### In the Air-Gapped Environment

```bash
# Deploy a registry
./capsailer registry --namespace my-registry

# Push all images from the bundle to the registry
./capsailer push --bundle capsailer-bundle.tar.gz --namespace my-registry

# Deploy applications using standard Helm commands
kubectl port-forward -n my-registry svc/chartmuseum 8080:8080 &
helm repo add local-charts http://localhost:8080
helm repo update
helm install my-redis local-charts/redis
```

## Using Values Files with Charts

This example shows how to include values files with Helm charts.

### Manifest File

```yaml
# manifest.yaml
images:
  - nginx:1.25
  - redis:7.0

charts:
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
    valuesFile: redis-values.yaml
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
    valuesFile: nginx-values.yaml
```

### Values Files

```yaml
# redis-values.yaml
master:
  persistence:
    size: 10Gi
replica:
  replicaCount: 2
```

```yaml
# nginx-values.yaml
service:
  type: NodePort
replicaCount: 3
```

### Build and Deploy

```bash
# Build the bundle
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz

# In the air-gapped environment
./capsailer push --bundle capsailer-bundle.tar.gz --namespace my-registry

# Deploy with values
helm install my-redis local-charts/redis -f redis-values.yaml
``` 