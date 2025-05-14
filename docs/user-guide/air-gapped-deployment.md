# Air-Gapped Deployment

This guide explains how to deploy applications in an air-gapped (offline) Kubernetes environment using Capsailer.

## Overview

Deploying in an air-gapped environment involves three main steps:

1. Transferring the bundle and Capsailer binary to the air-gapped environment
2. Setting up a local registry and chart repository
3. Pushing images and charts from the bundle to the registry

## Rewriting Image References

When deploying applications in air-gapped environments, you need to ensure that container image references in Helm charts point to your private registry. Capsailer provides a built-in feature to automatically rewrite these references during the build process:

```bash
# Build a bundle with image reference rewriting
capsailer build --manifest manifest.yaml --output bundle.tar.gz --rewrite-image-references --registry-url registry.local:5000
```

This will:
- Download all images and charts specified in the manifest
- Rewrite all image references in Helm charts to use your private registry
- Package everything into a portable bundle

When you deploy these charts in your air-gapped environment, they will automatically use images from your private registry without requiring any manual modifications.

## Handling Kubernetes Operators

Kubernetes operators often require special handling in air-gapped environments because:

1. Operators typically reference container images in their Custom Resources (CRs)
2. These image references need to be rewritten to point to your private registry

### Operator Image References in CRs

When deploying operators, you'll often need to specify container images in the Custom Resource (CR) definitions. For example, a database operator might require you to specify the database image in its CR:

```yaml
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: my-database
spec:
  # Image reference that needs to be rewritten
  image: docker.io/postgres:14.5
  replicas: 3
```

### Strategies for Handling Operators

1. **Include Operator Images in Your Manifest**:
   Make sure to include all images required by the operator in your Capsailer manifest.

2. **Manually Update CRs**:
   After deploying the operator, update the image references in your CRs to point to your private registry:

   ```yaml
   spec:
     # Updated to use private registry
     image: registry.local:5000/postgres:14.5
   ```

3. **Use Helm Values for Operators**:
   If deploying operators via Helm, use values files to override image references:

   ```yaml
   # values.yaml
   operator:
     image: registry.local:5000/operator:v1.0.0
   
   # Images used by the operator's CRs
   defaultImages:
     postgres: registry.local:5000/postgres:14.5
     redis: registry.local:5000/redis:7.0
   ```

4. **Leverage Capsailer's Image Rewriting**:
   If your operator is deployed via a Helm chart that includes CR templates, Capsailer's image reference rewriting feature will automatically update those references.

### Example: PostgreSQL Operator

```yaml
# In your manifest.yaml
images:
  - postgres:14.5
  - postgres-operator:v1.10.0
  # Any additional images the operator might need

charts:
  - name: postgres-operator
    repo: https://example.com/charts
    version: 1.10.0
```

After deploying with Capsailer:

```bash
# Deploy the operator
helm install postgres-operator local-charts/postgres-operator

# Create a database CR with the rewritten image reference
cat <<EOF | kubectl apply -f -
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: my-database
spec:
  image: registry.local:5000/postgres:14.5
  replicas: 3
EOF
```

## Transferring the Bundle

After building your bundle in a connected environment, you need to transfer it to the air-gapped environment:

```bash
# Copy the bundle and the capsailer binary
scp capsailer capsailer-bundle.tar.gz user@airgapped:~/
```

## Setting Up a Registry

In the air-gapped environment, you need to set up a local container registry and Helm chart repository:

```bash
# Deploy a registry in the my-registry namespace
./capsailer registry --namespace my-registry
```

This command deploys:

- A Docker registry for container images
- A ChartMuseum instance for Helm charts
- Persistent storage for both services (optional)

## Pushing Artifacts

Once the registry is set up, you can push all artifacts from your bundle to the registry:

```bash
# Push all images and charts from the bundle to the registry
./capsailer push --bundle capsailer-bundle.tar.gz --namespace my-registry
```

## Deploying Applications

After pushing the artifacts, you can deploy applications using standard Helm commands:

```bash
# First, add the ChartMuseum as a Helm repository
kubectl port-forward -n my-registry svc/chartmuseum 8080:8080 &
helm repo add local-charts http://localhost:8080
helm repo update

# Now install charts
helm install my-release local-charts/nginx --values values.yaml
```

Since the image references in the charts have been rewritten during the build process (if you used the `--rewrite-image-references` flag), the deployed applications will automatically use images from your private registry.

## Registry Options

You can customize the registry deployment:

```bash
# Deploy with custom settings
capsailer registry --namespace my-registry --image registry:2.8 --persistent=false
```

## Complete Workflow

Here's the complete workflow for air-gapped deployments:

1. **In the connected environment**:
   ```bash
   # Create a manifest file with your images and charts
   vi manifest.yaml
   
   # Validate the manifest
   capsailer init --manifest manifest.yaml
   
   # Build the bundle with image reference rewriting
   capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz --rewrite-image-references --registry-url registry.local:5000
   ```

2. **Transfer to air-gapped environment**:
   ```bash
   # Copy the bundle and the capsailer binary
   scp capsailer capsailer-bundle.tar.gz user@airgapped:~/
   ```

3. **In the air-gapped environment**:
   ```bash
   # Deploy a registry
   ./capsailer registry --namespace my-registry
   
   # Push all images from the bundle to the registry
   ./capsailer push --bundle capsailer-bundle.tar.gz --namespace my-registry
   
   # Deploy applications using standard Helm commands
   kubectl port-forward -n my-registry svc/chartmuseum 8080:8080 &
   helm repo add local-charts http://localhost:8080
   helm repo update
   helm install my-release local-charts/nginx
   ``` 