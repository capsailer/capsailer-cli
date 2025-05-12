# registry

The `registry` command deploys a standalone Docker registry and Helm chart repository in a Kubernetes cluster.

## Usage

```bash
capsailer registry [options]
```

## Description

The `registry` command deploys:

1. A Docker registry for container images
2. A ChartMuseum instance for Helm charts
3. Persistent storage for both services (optional)

This provides a simple way to set up a local registry for your air-gapped deployments.

## Options

| Option | Description |
|--------|-------------|
| `--namespace` | Kubernetes namespace to deploy the registry in (default: `default`) |
| `--image` | Docker image to use for the registry (default: `registry:2`) |
| `--persistent` | Whether to use persistent storage (default: `true`) |
| `--storage-class` | Storage class to use for persistent volumes |
| `--storage-size` | Size of the persistent volumes (default: `10Gi`) |
| `--kubeconfig` | Path to the kubeconfig file |
| `--port` | Port to expose the registry on (default: `5000`) |
| `--chart-port` | Port to expose the chart repository on (default: `8080`) |

## Examples

```bash
# Deploy a registry with default settings
capsailer registry

# Deploy a registry in a specific namespace
capsailer registry --namespace my-registry

# Deploy a registry with custom settings
capsailer registry --namespace my-registry --image registry:2.8 --persistent=false

# Deploy a registry with a specific kubeconfig
capsailer registry --kubeconfig /path/to/kubeconfig
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Failed to create namespace |
| 2 | Failed to deploy registry |
| 3 | Failed to deploy chart repository |

## See Also

- [Air-Gapped Deployment](../user-guide/air-gapped-deployment.md)
- [push](push.md)