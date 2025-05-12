# Registry Command

The `registry` command deploys a Docker registry and ChartMuseum repository in your Kubernetes cluster.

## Usage

```bash
capsailer registry [flags]
```

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--namespace` | Kubernetes namespace for the registry | `capsailer-registry` |
| `--image` | Container image for the registry | `registry:2` |
| `--persistent` | Whether to use persistent storage | `true` |
| `--kubeconfig` | Path to kubeconfig file | `~/.kube/config` |

## Description

The `registry` command sets up the necessary infrastructure for air-gapped deployments:

1. **Docker Registry** - For storing container images
2. **ChartMuseum** - For hosting Helm charts

This provides a complete solution for storing and accessing all artifacts needed for Kubernetes deployments in an air-gapped environment.

## Air-Gapped Considerations

When running in an air-gapped environment, the command automatically detects the lack of internet connectivity and provides guidance:

- Checks if the registry image is available locally in a bundle
- Provides options for handling the registry image in air-gapped scenarios
- Loads the registry image from a local bundle if available

## Examples

```bash
# Deploy with default settings
capsailer registry

# Deploy in a custom namespace
capsailer registry --namespace my-registry

# Deploy with ephemeral storage (no persistence)
capsailer registry --persistent=false

# Use a specific kubeconfig file
capsailer registry --kubeconfig /path/to/kubeconfig
```

## Workflow Integration

The `registry` command is typically the first step in an air-gapped deployment workflow:

```bash
# Step 1: Deploy registry infrastructure
capsailer registry --namespace my-registry

# Step 2: Push artifacts from a bundle
capsailer push --bundle capsailer-bundle.tar.gz --namespace my-registry

# Step 3: Deploy applications
capsailer deploy --chart nginx --registry-namespace my-registry
```

## Accessing the Registry

After deployment, the registry is available within the cluster at:

```
registry.<namespace>.svc.cluster.local:5000
```

ChartMuseum is available at:

```
chartmuseum.<namespace>.svc.cluster.local:8080
```

For external access, you can use port-forwarding:

```bash
# Port-forward the registry service
kubectl port-forward -n capsailer-registry svc/registry 5000:5000

# Port-forward the ChartMuseum service
kubectl port-forward -n capsailer-registry svc/chartmuseum 8080:8080
``` 