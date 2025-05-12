# init

The `init` command validates and normalizes a manifest file.

## Usage

```bash
capsailer init --manifest <manifest-file>
```

## Description

The `init` command performs the following actions:

1. Validates that the manifest file is properly formatted
2. Checks that all required fields are present
3. Normalizes image references (adds `latest` tag if missing)
4. Validates that chart references are properly formatted

## Options

| Option | Description |
|--------|-------------|
| `--manifest` | Path to the manifest file (required) |
| `--output` | Path to write the normalized manifest (optional) |

## Examples

```bash
# Validate a manifest file
capsailer init --manifest manifest.yaml

# Validate and write the normalized manifest to a new file
capsailer init --manifest manifest.yaml --output normalized-manifest.yaml
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Invalid manifest format |
| 2 | Missing required fields |

## See Also

- [Creating Manifests](../user-guide/creating-manifests.md)
- [build](build.md) 