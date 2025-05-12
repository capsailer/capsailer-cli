# Capsailer

Capsailer is a CLI tool for delivering Kubernetes applications into air-gapped (offline) environments.

## Overview

Capsailer allows you to define Helm charts and container images, package them into a portable archive, and deploy them into an air-gapped Kubernetes environment by installing a local registry and Helm chart server.

## Documentation

The full documentation is available at [https://jlnhnng.github.io/capsailer/](https://jlnhnng.github.io/capsailer/)

To build the documentation locally:

```bash
# Install dependencies
pip install -r requirements.txt

# Serve the documentation
mkdocs serve
```

## Features

- Download container images and Helm charts from public or private repositories
- Package everything into a single, portable archive file
- Deploy the bundle in an air-gapped environment
- Set up a local container registry and Helm chart repository
- Self-contained CLI that doesn't require Docker or skopeo for image operations
- Built-in support for pushing container images and Helm charts without external dependencies
- Automated chart repository deployment and publishing

## How It Works

Capsailer provides an all-in-one solution for air-gapped Kubernetes deployments:

1. **Bundle Creation**: Package container images and Helm charts into a portable bundle
2. **Registry Infrastructure**: Deploy a container registry and Helm chart repository
3. **Push Mechanism**: Upload images and charts without requiring external tools like Docker or skopeo

Unlike other solutions, Capsailer handles both container images and Helm charts natively, without relying on external tools in the air-gapped environment.

## Prerequisites

- Go 1.20 or later
- Kubernetes cluster (for deployment)
- Admin access to the cluster

## Installation

```bash
# Clone the repository
git clone https://github.com/jlnhnng/capsailer.git
cd capsailer

# Build the binary
go build -o capsailer cmd/capsailer/main.go

# Add to your PATH
mv capsailer /usr/local/bin/
```

## Usage

### Creating a Manifest

Create a manifest file that describes the images and charts you want to include:

```yaml
images:
  - nginx:1.25
  - redis:7.0
  - bitnami/postgresql:15.4.0

charts:
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
    valuesFile: redis-values.yaml
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
```

### Building a Bundle

```bash
capsailer init --manifest manifest.yaml
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz
```

### Deploying in an Air-Gapped Environment

```bash
# Copy the bundle to the air-gapped environment
scp capsailer-bundle.tar.gz user@airgapped:~/

# On the air-gapped system
# Step 1: Deploy a registry
capsailer registry --namespace my-registry

# Step 2: Push all images from the bundle to the registry
capsailer push --bundle capsailer-bundle.tar.gz --namespace my-registry

```

## Command Reference

- `capsailer init`: Validate and normalize the manifest
- `capsailer build`: Download and package images and charts
- `capsailer unpack`: Extract bundle and set up local registry
- `capsailer registry`: Deploy a standalone Docker registry in a Kubernetes cluster
- `capsailer push`: Push container images to the registry

### Complete Workflow for Air-Gapped Deployments

Here's the complete workflow for air-gapped deployments:

1. **In the connected environment**:
   ```bash
   # Create a manifest file with your images and charts
   vi manifest.yaml
   
   # Validate the manifest
   capsailer init --manifest manifest.yaml
   
   # Build the bundle
   capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz
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
   # First, add the ChartMuseum as a Helm repository
   kubectl port-forward -n my-registry svc/chartmuseum 8080:8080 &
   helm repo add local-charts http://localhost:8080
   helm repo update
   
   # Now install charts
   helm install my-release local-charts/nginx --values values.yaml
   ```

### Using the Registry Command

You can deploy a standalone Docker registry to your Kubernetes cluster:

```bash
# Deploy with default settings
capsailer registry

# Deploy with custom settings
capsailer registry --namespace my-registry --image registry:2.8 --kubeconfig ~/.kube/config --persistent=false
```

This provides a simple way to set up a local registry for your air-gapped deployments without going through the full unpack process.

### Using the Push Command

The push command allows you to upload both container images and Helm charts to the registry and chart repository:

```bash
# Push a single image to the registry
capsailer push --image nginx:latest --namespace my-registry

# Push all images and charts from a bundle to the registry
capsailer push --bundle capsailer-bundle.tar.gz --namespace my-registry

# Push artifacts from an unpacked bundle directory
capsailer push --bundle ./unpacked-bundle --namespace my-registry

# Push to an external registry (like Artifactory)
capsailer push --bundle capsailer-bundle.tar.gz --external-registry artifactory.example.com --username myuser --password mypassword
```

This command handles all the necessary steps:
1. Finding the registry service in the specified namespace (or using the provided external registry)
2. Setting up a Helm chart repository if needed (for internal registry only)
3. Loading images from the bundle without requiring Docker or skopeo
4. Pushing images directly to the registry using built-in container registry library
5. Publishing Helm charts to the chart repository using direct HTTP API calls (for internal registry only)

Unlike many similar tools, Capsailer doesn't rely on external dependencies like Docker or skopeo to push images and charts, making it truly self-contained and perfect for air-gapped environments.

### Helm Chart Management

Capsailer provides comprehensive support for Helm charts:

1. **Chart Downloading**: Download Helm charts from any repository during bundle creation
2. **Chart Repository**: Automatically deploy a Chartmuseum instance to host charts
3. **Chart Publishing**: Upload Helm charts to the repository without external tools
4. **Chart Deployment**: Deploy applications using the locally hosted charts

This end-to-end chart handling makes Capsailer ideal for managing complex applications in air-gapped environments.

#### Air-Gapped Bundle Deployment Workflow

In an air-gapped environment, Capsailer sets up a complete infrastructure for deploying applications:

1. **Docker Registry**: For storing container images
2. **Chart Repository**: For hosting Helm charts

All artifacts from your bundle (images and charts) are pushed to these repositories, making them available for deployment in your air-gapped Kubernetes cluster.

### Air-Gapped Registry Deployment

When deploying a registry in an air-gapped environment, Capsailer offers several strategies:

1. **Bundle-Included Registry Image**: If you've included the `registry:2` image in your bundle (recommended):
   ```bash
   # First build a bundle with the registry image
   capsailer build --manifest manifest.yaml --output bundle.tar.gz

   # Then unpack the bundle in the air-gapped environment
   capsailer unpack --file bundle.tar.gz

   # Now deploy the registry using the locally available image
   capsailer registry --image localhost:5000/registry:2
   ```

2. **Local Build**: Build a registry image locally (requires Docker):
   ```bash
   # This will build a registry image locally and deploy it
   capsailer registry --local-build
   ```

3. **Pre-loaded Registry Image**: If the registry image is pre-loaded on your cluster nodes:
   ```bash
   # No need for internet access
   capsailer registry
   ```

4. **Manual Image Transfer**: If you have the registry image available as a tar file:
   ```bash
   # Capsailer will detect the air-gapped environment
   # and load the image from the local filesystem
   capsailer registry
   ```

Capsailer will automatically detect if you're in an air-gapped environment and provide appropriate guidance.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 