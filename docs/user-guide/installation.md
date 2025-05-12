# Installation

There are several ways to install Capsailer on your system.

## Building from Source

```bash
# Clone the repository
git clone https://github.com/jlnhnng/capsailer.git
cd capsailer

# Build the binary
go build -o capsailer cmd/capsailer/main.go

# Add to your PATH
mv capsailer /usr/local/bin/
```

## Prerequisites

Capsailer requires:

- Go 1.20 or later (for building from source)
- Kubernetes cluster (for deployment)
- Admin access to the cluster

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