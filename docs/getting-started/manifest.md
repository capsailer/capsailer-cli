# Manifest File

The manifest file is a YAML document that defines what container images and Helm charts you want to include in your bundle. This page explains the manifest file format and provides examples.

## Basic Structure

A manifest file has two main sections:

1. `images` - A list of container images to include
2. `charts` - A list of Helm charts to include

Here's a basic example:

```yaml
images:
  - nginx:1.25.0
  - redis:7.0.14
  - bitnami/postgresql:15.4.0

charts:
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
    valuesFile: redis-values.yaml
```

## Images Section

The `images` section is a simple list of container images in the format `repository:tag`. For example:

```yaml
images:
  - nginx:1.25.0           # From Docker Hub
  - redis:7.0.14           # From Docker Hub
  - bitnami/postgresql:15.4.0  # From Docker Hub with organization
  - registry:2             # Registry image (recommended for air-gapped deployments)
  - ghcr.io/user/app:v1    # From GitHub Container Registry
  - quay.io/company/app:latest  # From Quay.io
```

### Best Practices for Images

1. **Always specify tags** - Avoid using `latest` as it can lead to inconsistencies
2. **Include the registry image** - Always include `registry:2` for air-gapped deployments
3. **Pin specific versions** - Use exact versions for reproducible deployments
4. **Include all dependencies** - Include any sidecar or init container images your application needs

## Charts Section

The `charts` section is a list of objects, each with the following properties:

| Property | Description | Required |
|----------|-------------|----------|
| `name` | Chart name | Yes |
| `repo` | Chart repository URL | Yes |
| `version` | Chart version | Yes |
| `valuesFile` | Path to a values file (optional) | No |

Example:

```yaml
charts:
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
  
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
    valuesFile: redis-values.yaml
  
  - name: postgresql
    repo: https://charts.bitnami.com/bitnami
    version: 12.5.7
```

### Values Files

If you specify a `valuesFile`, Capsailer will:

1. Include the values file in the bundle
2. Use it when deploying the chart
3. Automatically rewrite image references to use the local registry

Example values file (redis-values.yaml):

```yaml
architecture: standalone
auth:
  enabled: false
master:
  persistence:
    enabled: false
```

### Best Practices for Charts

1. **Specify exact versions** - Always pin chart versions for consistency
2. **Use values files** - Customize charts with values files rather than editing the charts directly
3. **Include dependencies** - If your application depends on other services, include their charts

## Complete Example

Here's a complete example of a manifest file for a web application with a database:

```yaml
images:
  # Include the registry image for air-gapped deployments
  - registry:2
  
  # Application images
  - nginx:1.25.0
  - bitnami/postgresql:15.4.0
  - redis:7.0.14
  
  # Include any sidecar containers
  - fluent/fluentd:v1.16.2
  - jaegertracing/jaeger-agent:1.47.0

charts:
  # Main application charts
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
    valuesFile: nginx-values.yaml
  
  - name: postgresql
    repo: https://charts.bitnami.com/bitnami
    version: 12.5.7
    valuesFile: postgresql-values.yaml
  
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
    valuesFile: redis-values.yaml
```

## Validating the Manifest

You can validate your manifest using the `init` command:

```bash
capsailer init --manifest manifest.yaml
```

This will:
1. Check that the manifest is valid YAML
2. Verify that all required fields are present
3. Ensure that values files exist
4. Output a summary of the images and charts included

## Next Steps

Now that you understand the manifest file format, you can:

1. Create a manifest for your application
2. Follow the [quick start guide](quick-start.md) to build and deploy your bundle
3. Learn about [air-gapped environments](../concepts/air-gapped.md) 