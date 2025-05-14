# build

The `build` command downloads container images and Helm charts and packages them into a portable bundle.

## Usage

```bash
capsailer build --manifest <manifest-file> --output <output-file>
```

## Description

The `build` command performs the following actions:

1. Reads the manifest file
2. Downloads all container images specified in the manifest
3. Saves the images as OCI artifacts
4. Downloads all Helm charts specified in the manifest
5. Optionally rewrites image references in Helm charts to use a private registry
6. Packages everything into a single, portable archive file

## Options

| Option | Description |
|--------|-------------|
| `--manifest` | Path to the manifest file (required) |
| `--output` | Path to write the bundle file (required) |
| `--rewrite-image-references` | Rewrite image references in Helm charts to use a private registry |
| `--registry-url` | URL of the private registry to use when rewriting image references |
| `--username` | Username for authentication with private registries |
| `--password` | Password for authentication with private registries |
| `--kubeconfig` | Path to the kubeconfig file |
| `--registry-url` | URL of the registry to use for image pulls |
| `--skip-tls-verify` | Skip TLS verification when pulling images |

## Examples

```bash
# Build a bundle from a manifest
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz

# Build a bundle with image reference rewriting
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz --rewrite-image-references --registry-url registry.local:5000

# Build a bundle with authentication for private registries
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz --username myuser --password mypassword

# Build a bundle with a specific kubeconfig
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz --kubeconfig /path/to/kubeconfig
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Failed to read manifest |
| 2 | Failed to download images |
| 3 | Failed to download charts |
| 4 | Failed to create bundle |

## See Also

- [Building Bundles](../user-guide/building-bundles.md)
- [init](init.md)
- [push](push.md) 