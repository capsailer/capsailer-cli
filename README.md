# Capsailer

Capsailer is a CLI tool for delivering Kubernetes applications into air-gapped (offline) environments.

[![Documentation](https://img.shields.io/badge/docs-capsailer.dev-blue)](https://docs.capsailer.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

## Overview

Capsailer packages Helm charts and container images into a portable archive that can be deployed in air-gapped Kubernetes environments. It handles the entire workflow from bundle creation to deployment without requiring external dependencies like Docker or skopeo.

## Key Features

- Package container images and Helm charts into a single portable bundle
- Deploy a local container registry and Helm chart repository in air-gapped environments
- Push images and charts without requiring external tools
- Self-contained CLI with no dependencies in the air-gapped environment

## Installation

### Download Pre-built Binary

```bash
# Linux (amd64)
curl -Lo capsailer.tar.gz https://github.com/capsailer/capsailer-cli/releases/latest/download/capsailer-linux-amd64.tar.gz
tar -xzf capsailer.tar.gz
chmod +x capsailer
sudo mv capsailer /usr/local/bin/

# macOS (Intel)
curl -Lo capsailer.tar.gz https://github.com/capsailer/capsailer-cli/releases/latest/download/capsailer-darwin-amd64.tar.gz
tar -xzf capsailer.tar.gz
chmod +x capsailer
sudo mv capsailer /usr/local/bin/

# macOS (Apple Silicon)
curl -Lo capsailer.tar.gz https://github.com/capsailer/capsailer-cli/releases/latest/download/capsailer-darwin-arm64.tar.gz
tar -xzf capsailer.tar.gz
chmod +x capsailer
sudo mv capsailer /usr/local/bin/
```

For Windows, download from the [releases page](https://github.com/capsailer/capsailer-cli/releases/latest).

### Build from Source

```bash
git clone https://github.com/capsailer/capsailer-cli.git
cd capsailer
go build -o capsailer cmd/capsailer/main.go
```

## Quick Start

### 1. Create a Manifest

```yaml
images:
  - nginx:1.25
  - redis:7.0

charts:
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
    valuesFile: redis-values.yaml
```

### 2. Build a Bundle (Connected Environment)

```bash
capsailer init --manifest manifest.yaml
capsailer build --manifest manifest.yaml --output bundle.tar.gz
```

### 3. Deploy in Air-Gapped Environment

```bash
# Deploy registry
capsailer registry --namespace my-registry

# Push artifacts from bundle
capsailer push --bundle bundle.tar.gz --namespace my-registry

# Deploy applications
kubectl port-forward -n my-registry svc/chartmuseum 8080:8080 &
helm repo add local-charts http://localhost:8080
helm repo update
helm install my-release local-charts/redis
```

## Command Reference

| Command | Description |
|---------|-------------|
| `init` | Validate and normalize the manifest |
| `build` | Download and package images and charts |
| `registry` | Deploy a container registry in Kubernetes |
| `push` | Push images and charts to a registry |
| `unpack` | Extract bundle contents |

## Documentation

For complete documentation, visit [docs.capsailer.dev](https://docs.capsailer.dev/) or build locally:

```bash
pip install -r requirements.txt
mkdocs serve
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 