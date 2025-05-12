# Push Command

The `push` command uploads container images and Helm charts to the registry in your Kubernetes cluster.

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

## Description

The `push` command handles uploading artifacts to the local infrastructure:

1. **Image Pushing** - Uploads container images to the Docker registry
2. **Chart Publishing** - Publishes Helm charts to ChartMuseum

This command works without requiring external tools like Docker or skopeo, making it ideal for air-gapped environments where such tools might not be available.

## Features

- **Self-contained** - No dependencies on Docker or other external tools
- **Bundle Support** - Push all artifacts from a bundle in one command
- **Single Image Mode** - Push individual images when needed
- **Automatic Discovery** - Finds the registry and ChartMuseum services in the cluster
- **Direct Registry API** - Uses direct registry API calls for pushing images
- **ChartMuseum Integration** - Automatically publishes charts to ChartMuseum

## Examples

```bash
# Push all artifacts from a bundle
capsailer push --bundle capsailer-bundle.tar.gz

# Push a single image
capsailer push --image nginx:latest
```

## Workflow Integration

The `push` command is typically used after setting up the registry and before deploying applications:

```