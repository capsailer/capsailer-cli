# Getting Started

This guide will help you get started with Capsailer quickly.

## Prerequisites

- Kubernetes cluster (for deployment)
- Admin access to the cluster

## Installation

The quickest way to get started is to download a pre-built release:

```bash
# For Linux (amd64)
curl -Lo capsailer https://github.com/jlnhnng/capsailer/releases/latest/download/capsailer-linux-amd64.tar.gz
tar -xzf capsailer-linux-amd64.tar.gz
chmod +x capsailer

# For macOS (Intel)
curl -Lo capsailer https://github.com/jlnhnng/capsailer/releases/latest/download/capsailer-darwin-amd64.tar.gz
tar -xzf capsailer-darwin-amd64.tar.gz
chmod +x capsailer

# For macOS (Apple Silicon)
curl -Lo capsailer https://github.com/jlnhnng/capsailer/releases/latest/download/capsailer-darwin-arm64.tar.gz
tar -xzf capsailer-darwin-arm64.tar.gz
chmod +x capsailer

# For Windows
# Download from https://github.com/jlnhnng/capsailer/releases/latest

# Add to your PATH
mv capsailer /usr/local/bin/
```

You can find all available releases at: https://github.com/jlnhnng/capsailer/releases

## Quick Start

### 1. Create a Manifest

Create a manifest file that describes the images and charts you want to include:

```yaml
images:
  - nginx:1.25
  - redis:7.0

charts:
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
```

### 2. Build a Bundle

```bash
capsailer init --manifest manifest.yaml
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz
```

### 3. Deploy in an Air-Gapped Environment

```bash
# Deploy a registry
capsailer registry --namespace my-registry

# Push all images from the bundle to the registry
capsailer push --bundle capsailer-bundle.tar.gz --namespace my-registry
```

## Next Steps

- Learn more about [creating manifests](user-guide/creating-manifests.md)
- Explore the [command reference](commands/overview.md)
- Check out the [examples](examples.md) 