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
4. Optionally rewrites image references in Helm charts to use a private registry
5. Packages everything into a single, portable archive file

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

## Rewriting Image References

For air-gapped deployments, you often need to rewrite container image references in Helm charts to point to your private registry. Capsailer can do this automatically during the build process:

```bash
# Build a bundle with image reference rewriting
capsailer build --manifest manifest.yaml --output bundle.tar.gz --rewrite-image-references --registry-url registry.local:5000
```

This will:
- Download all images and charts specified in the manifest
- Rewrite all image references in Helm charts to use your private registry
- Package everything into a portable bundle

When you deploy these charts in your air-gapped environment, they will automatically use images from your private registry without requiring any manual modifications.

## Including Operator Images

When building bundles that include Kubernetes operators, you need to consider both the operator images themselves and the images referenced in the operator's Custom Resources (CRs):

```yaml
# Example manifest.yaml for an operator
images:
  # The operator image itself
  - quay.io/example/postgres-operator:v1.10.0
  
  # Images that the operator will deploy via CRs
  - docker.io/postgres:14.5
  - docker.io/postgres:14.6
  - docker.io/postgres-exporter:0.10.0

charts:
  - name: postgres-operator
    repo: https://example.com/charts
    version: 1.10.0
```

### Tips for Operator Bundles

1. **Identify All Required Images**: Review the operator documentation to identify all container images that might be deployed by the operator's CRs.

2. **Include Related Tools**: Many operators deploy additional components like exporters, sidecars, or init containers. Make sure to include these images in your manifest.

3. **Version Consistency**: Ensure that the versions of the operator and the images it deploys are compatible with each other.

4. **Check CR Templates**: If the operator's Helm chart includes CR templates, Capsailer's image reference rewriting will automatically update those references.

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