# Installation

This guide covers different ways to install Capsailer.

## Pre-built Binaries

The easiest way to install Capsailer is to download a pre-built release from the [GitHub Releases page](https://github.com/capsailer/capsailer-cli/releases).

### Linux (amd64)

```bash
# Download the binary
curl -Lo capsailer.tar.gz https://github.com/capsailer/capsailer-cli/releases/latest/download/capsailer-linux-amd64.tar.gz
tar -xzf capsailer.tar.gz
chmod +x capsailer
sudo mv capsailer /usr/local/bin/
```

### Linux (arm64)

```bash
# Download the binary
curl -Lo capsailer.tar.gz https://github.com/capsailer/capsailer-cli/releases/latest/download/capsailer-linux-arm64.tar.gz
tar -xzf capsailer.tar.gz
chmod +x capsailer
sudo mv capsailer /usr/local/bin/
```

### macOS (Intel)

```bash
# Download the binary
curl -Lo capsailer.tar.gz https://github.com/capsailer/capsailer-cli/releases/latest/download/capsailer-darwin-amd64.tar.gz
tar -xzf capsailer.tar.gz
chmod +x capsailer
sudo mv capsailer /usr/local/bin/
```

### macOS (Apple Silicon)

```bash
# Download the binary
curl -Lo capsailer.tar.gz https://github.com/capsailer/capsailer-cli/releases/latest/download/capsailer-darwin-arm64.tar.gz
tar -xzf capsailer.tar.gz
chmod +x capsailer
sudo mv capsailer /usr/local/bin/
```

### Windows

1. Download the Windows binary from the [GitHub Releases page](https://github.com/capsailer/capsailer-cli/releases/latest)
2. Extract the ZIP file
3. Add the extracted directory to your PATH

## Building from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/capsailer/capsailer-cli.git
cd capsailer-cli

# Build the binary
go build -o capsailer cmd/capsailer/main.go

# Add to your PATH
sudo mv capsailer /usr/local/bin/
```

## Prerequisites

Capsailer requires:

- Kubernetes cluster (for deployment)
- Admin access to the cluster
- Go 1.20 or later (only if building from source)

## Verifying Installation

After installation, verify that Capsailer is working correctly:

```bash
capsailer --version
```

You should see the current version of Capsailer displayed.

## Environment Setup

Capsailer uses your Kubernetes configuration by default. Make sure your `kubectl` is properly configured to connect to your cluster:

```bash
kubectl config current-context
```

If you need to use a specific kubeconfig file, you can specify it with the `--kubeconfig` flag:

```bash
capsailer registry --kubeconfig /path/to/kubeconfig
``` 