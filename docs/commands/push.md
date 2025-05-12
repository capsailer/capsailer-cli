# Push Command

The `push` command uploads container images and Helm charts to the registry in your Kubernetes cluster or an external registry.

## Usage

```bash
capsailer push [flags]
```

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--bundle` | Path to the bundle file or directory | |
| `--image` | Specific image to push | |
| `--namespace` | Kubernetes namespace for the registry | `capsailer-registry` |
| `--kubeconfig` | Path to kubeconfig file | `~/.kube/config` |
| `--external-registry` | External registry URL to push images to | |
| `--username` | Username for authentication with external registry | |
| `--password` | Password for authentication with external registry | |

## Description

The `push` command handles uploading artifacts to the local or external infrastructure:

1. **Image Pushing** - Uploads container images to the Docker registry
2. **Chart Publishing** - Publishes Helm charts to ChartMuseum (internal registry only)

This command works without requiring external tools like Docker or skopeo, making it ideal for air-gapped environments where such tools might not be available.

## Features

- **Self-contained** - No dependencies on Docker or other external tools
- **Bundle Support** - Push all artifacts from a bundle in one command
- **Single Image Mode** - Push individual images when needed
- **Automatic Discovery** - Finds the registry and ChartMuseum services in the cluster
- **Direct Registry API** - Uses direct registry API calls for pushing images
- **ChartMuseum Integration** - Automatically publishes charts to ChartMuseum
- **External Registry Support** - Push to external registries like Artifactory or Docker Hub

## Examples

```bash
# Push all artifacts from a bundle to the internal registry
capsailer push --bundle capsailer-bundle.tar.gz

# Push a single image to the internal registry
capsailer push --image nginx:latest

# Push all images from a bundle to an external registry
capsailer push --bundle capsailer-bundle.tar.gz --external-registry artifactory.example.com --username myuser --password mypassword
```

## Workflow Integration

The `push` command is typically used after setting up the registry and before deploying applications:

```bash
# Set up registry (for internal registry)
capsailer registry --namespace my-registry

# Push artifacts
capsailer push --bundle capsailer-bundle.tar.gz --namespace my-registry

# Or push to an external registry
capsailer push --bundle capsailer-bundle.tar.gz --external-registry artifactory.example.com --username myuser --password mypassword

# Deploy applications
# For internal registry:
kubectl port-forward -n my-registry svc/chartmuseum 8080:8080 &
helm repo add local-charts http://localhost:8080
helm install my-app local-charts/my-app

# For external registry:
# Use standard Helm commands with the external registry URL
helm install my-app --set image.registry=artifactory.example.com my-app
```

## Notes on External Registries

When pushing to external registries:

1. Helm charts are not published (only container images are pushed)
2. Authentication is supported via username/password
3. Docker credentials from ~/.docker/config.json are used if available
4. The external registry must support the Docker Registry API v2