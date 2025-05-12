# Getting Started

This guide will help you get started with Capsailer quickly.

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