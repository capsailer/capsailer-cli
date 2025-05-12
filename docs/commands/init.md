# Init Command

The `init` command validates and normalizes a manifest file.

## Usage

```bash
capsailer init --manifest MANIFEST_FILE [flags]
```

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--manifest` | Path to the manifest file | `manifest.yaml` |

## Examples

```bash
# Validate the default manifest file
capsailer init

# Validate a specific manifest file
capsailer init --manifest my-manifest.yaml
```

## Description

The `init` command performs several important validation steps:

1. **YAML Validation** - Ensures the manifest file is valid YAML
2. **Schema Validation** - Verifies that all required fields are present and properly formatted
3. **Values File Verification** - Checks that any referenced values files exist
4. **Summary Generation** - Provides a summary of the images and charts included in the manifest

This command is typically run before the `build` command to catch any issues with the manifest file early in the process.

## Output

The command outputs a summary of the manifest contents:

```
Initializing manifest from manifest.yaml
Manifest is valid. Found 3 images and 2 charts.
```

If there are any issues with the manifest, the command will provide detailed error messages:

```
Error: failed to load manifest: chart 'nginx' is missing required field 'repo'
```

## Workflow Integration

The `init` command is typically the first step in the workflow:

```bash
# Step 1: Initialize and validate the manifest
capsailer init --manifest manifest.yaml

# Step 2: Build the bundle
capsailer build --manifest manifest.yaml --output bundle.tar.gz
```

## Manifest File Format

The manifest file should follow this structure:

```yaml
images:
  - nginx:1.25.0
  - redis:7.0.14
  - registry:2

charts:
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
    valuesFile: redis-values.yaml
```

For more details on the manifest file format, see the [Manifest File](../getting-started/manifest.md) documentation. 