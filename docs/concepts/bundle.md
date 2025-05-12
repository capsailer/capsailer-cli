# Bundle Structure

Capsailer bundles are portable archives that contain all the artifacts needed for deploying applications in air-gapped environments. This page explains the structure and contents of these bundles.

## Overview

A Capsailer bundle is a compressed tar archive (`.tar.gz`) that contains:

1. Container images
2. Helm charts
3. Metadata
4. Manifest file

This self-contained package allows for easy transfer to air-gapped environments and provides everything needed for deployment without internet access.

## Directory Structure

When a bundle is unpacked, it creates the following directory structure:

```
capsailer-bundle/
├── images/
│   ├── nginx_1.25.0.tar
│   ├── redis_7.0.14.tar
│   └── ...
├── charts/
│   ├── nginx-15.1.4.tgz
│   ├── redis-17.11.7.tgz
│   └── ...
├── manifest.yaml
└── metadata.json
```

### Images Directory

The `images/` directory contains container images saved as tar archives. Each image is named according to its repository and tag, with slashes replaced by underscores:

- `nginx_1.25.0.tar` - From `nginx:1.25.0`
- `bitnami_postgresql_15.4.0.tar` - From `bitnami/postgresql:15.4.0`

These tar archives contain the full image, including all layers and metadata, ready to be loaded into a container registry.

### Charts Directory

The `charts/` directory contains Helm charts as `.tgz` archives. Each chart is named according to its name and version:

- `nginx-15.1.4.tgz`
- `redis-17.11.7.tgz`

These are standard Helm chart packages that can be installed directly or published to a chart repository.

### Manifest File

The `manifest.yaml` file is a copy of the original manifest used to build the bundle. It contains:

- List of container images
- List of Helm charts with their sources and versions
- Any custom values files referenced

Example:

```yaml
images:
  - nginx:1.25.0
  - redis:7.0.14
charts:
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
    valuesFile: nginx-values.yaml
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
```

### Metadata File

The `metadata.json` file contains information about the bundle itself:

```json
{
  "created": "2023-06-15T10:30:45Z",
  "version": "1.0.0",
  "capsailerVersion": "0.1.0",
  "images": [
    {
      "name": "nginx",
      "tag": "1.25.0",
      "digest": "sha256:abcd1234...",
      "size": 142586400
    },
    ...
  ],
  "charts": [
    {
      "name": "nginx",
      "version": "15.1.4",
      "size": 45872
    },
    ...
  ]
}
```

## Working with Bundles

### Building a Bundle

Bundles are created using the `build` command:

```bash
capsailer build --manifest manifest.yaml --output my-bundle.tar.gz
```

This downloads all required images and charts, and packages them into a single archive.

### Unpacking a Bundle

Bundles can be unpacked using the `unpack` command:

```bash
capsailer unpack --file my-bundle.tar.gz
```

This extracts the contents to the current directory, creating the directory structure described above.

### Pushing Bundle Contents

The contents of a bundle can be pushed to a registry using the `push` command:

```bash
capsailer push --bundle my-bundle.tar.gz
```

This uploads all images to the Docker registry and all charts to ChartMuseum.

## Bundle Size Considerations

Bundles can become quite large, especially when including multiple container images. Some strategies to manage bundle size:

1. **Selective Inclusion**: Only include the images and charts you need
2. **Layer Sharing**: Use images with shared base layers to reduce duplication
3. **Multiple Bundles**: Split large applications into multiple bundles
4. **Compression**: Bundles are compressed to reduce transfer size

## Security Aspects

Bundles should be treated as sensitive artifacts:

- They contain the complete application stack
- May include proprietary or sensitive code
- Should be verified for integrity after transfer
- Consider using encryption for transfer to air-gapped environments 