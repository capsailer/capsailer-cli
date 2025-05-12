# Building Bundles

Once you have created a manifest file, you can build a bundle that contains all the container images and Helm charts specified in the manifest.

## Basic Bundle Creation

```bash
# First, validate your manifest
capsailer init --manifest manifest.yaml

# Then build the bundle
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz
```

## What Happens During the Build

When you run the `build` command, Capsailer:

1. Downloads all container images specified in the manifest
2. Saves the images as OCI artifacts
3. Downloads all Helm charts specified in the manifest
4. Packages everything into a single, portable archive file

## Build Options

The `build` command supports several options:

```bash
# Specify an output file
capsailer build --manifest manifest.yaml --output my-bundle.tar.gz

# Use authentication for private registries
capsailer build --manifest manifest.yaml --username myuser --password mypassword

# Use a specific kubeconfig file
capsailer build --manifest manifest.yaml --kubeconfig /path/to/kubeconfig
```

## Bundle Contents

A Capsailer bundle contains:

- Container images in OCI format
- Helm charts
- A copy of the manifest file
- Metadata about the bundle

## Examining a Bundle

You can examine the contents of a bundle without extracting it:

```bash
capsailer inspect --bundle capsailer-bundle.tar.gz
```

This will show you a list of all the images and charts included in the bundle.

## Bundle Size Considerations

Bundle size depends on the number and size of the container images and Helm charts included. To keep bundle sizes manageable:

- Only include the specific images and charts you need
- Use specific tags rather than `latest` to avoid downloading unnecessary updates
- Consider using smaller base images when possible 