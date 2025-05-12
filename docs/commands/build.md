# Build Command

The `build` command downloads container images and Helm charts, and packages them into a portable bundle.

## Usage

```bash
capsailer build --manifest MANIFEST_FILE --output OUTPUT_FILE
```

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--manifest` | Path to the manifest file | `manifest.yaml` |
| `--output` | Output file path | `capsailer-bundle.tar.gz` |

## Example

```bash
# Build a bundle using the default manifest file
capsailer build --output my-bundle.tar.gz

# Build a bundle using a specific manifest file
capsailer build --manifest my-manifest.yaml --output my-bundle.tar.gz
``` 