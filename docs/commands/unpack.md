# unpack

The `unpack` command extracts a bundle and sets up a local registry.

## Usage

```bash
capsailer unpack --bundle <bundle-file> [options]
```

## Description

The `unpack` command performs the following actions:

1. Extracts the bundle to a directory
2. Sets up a local registry if requested
3. Prepares the extracted artifacts for use

This command is useful when you want to inspect the contents of a bundle or set up a local environment for testing.

## Options

| Option | Description |
|--------|-------------|
| `--bundle` | Path to the bundle file (required) |
| `--output` | Directory to extract the bundle to (default: `./unpacked-bundle`) |
| `--setup-registry` | Whether to set up a local registry (default: `false`) |
| `--registry-port` | Port to expose the registry on (default: `5000`) |
| `--chart-port` | Port to expose the chart repository on (default: `8080`) |

## Examples

```bash
# Extract a bundle to the default directory
capsailer unpack --bundle capsailer-bundle.tar.gz

# Extract a bundle to a specific directory
capsailer unpack --bundle capsailer-bundle.tar.gz --output ./my-bundle

# Extract a bundle and set up a local registry
capsailer unpack --bundle capsailer-bundle.tar.gz --setup-registry

# Extract a bundle and set up a local registry with custom ports
capsailer unpack --bundle capsailer-bundle.tar.gz --setup-registry --registry-port 5001 --chart-port 8081
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Failed to extract bundle |
| 2 | Failed to set up registry |

## See Also

- [Building Bundles](../user-guide/building-bundles.md)
- [Air-Gapped Deployment](../user-guide/air-gapped-deployment.md) 