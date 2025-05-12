# push

The `push` command uploads container images and Helm charts from a bundle to a registry.

## Usage

```bash
capsailer push --bundle <bundle-file> [options]
```

## Description

The `push` command performs the following actions:

1. Finds the registry service in the specified namespace (or uses the provided external registry)
2. Sets up a Helm chart repository if needed (for internal registry only)
3. Loads images from the bundle without requiring Docker or skopeo
4. Pushes images directly to the registry using built-in container registry library
5. Publishes Helm charts to the chart repository using direct HTTP API calls (for internal registry only)

Unlike many similar tools, Capsailer doesn't rely on external dependencies like Docker or skopeo to push images and charts, making it truly self-contained and perfect for air-gapped environments.

## Options

| Option | Description |
|--------|-------------|
| `--bundle` | Path to the bundle file or directory (required) |
| `--namespace` | Kubernetes namespace where the registry is deployed |
| `--external-registry` | URL of an external registry to push to |
| `--username` | Username for authentication with the registry |
| `--password` | Password for authentication with the registry |
| `--kubeconfig` | Path to the kubeconfig file |
| `--skip-tls-verify` | Skip TLS verification when pushing to the registry |
| `--image` | Push only a specific image from the bundle |

## Examples

```bash
# Push all images and charts from a bundle to the registry
capsailer push --bundle capsailer-bundle.tar.gz --namespace my-registry

# Push artifacts from an unpacked bundle directory
capsailer push --bundle ./unpacked-bundle --namespace my-registry

# Push to an external registry
capsailer push --bundle capsailer-bundle.tar.gz --external-registry artifactory.example.com --username myuser --password mypassword

# Push a single image to the registry
capsailer push --image nginx:latest --namespace my-registry
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Failed to load bundle |
| 2 | Failed to find registry |
| 3 | Failed to push images |
| 4 | Failed to push charts |

## See Also

- [Air-Gapped Deployment](../user-guide/air-gapped-deployment.md)
- [registry](registry.md) 