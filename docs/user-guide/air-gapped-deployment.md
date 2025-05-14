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