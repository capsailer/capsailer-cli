# Creating Manifests

A manifest file is a YAML file that defines the container images and Helm charts you want to include in your bundle.

## Manifest Structure

A basic manifest file looks like this:

```yaml
images:
  - nginx:1.25
  - redis:7.0
  - bitnami/postgresql:15.4.0

charts:
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
    valuesFile: redis-values.yaml
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
```

## Images Section

The `images` section is a list of container images you want to include in your bundle. Each image is specified in the format `repository:tag`.

```yaml
images:
  - nginx:1.25           # Public Docker Hub image
  - redis:7.0            # Public Docker Hub image
  - bitnami/postgresql:15.4.0  # Public Docker Hub image with namespace
  - registry.example.com/app:latest  # Private registry image
```

## Charts Section

The `charts` section is a list of Helm charts you want to include in your bundle. Each chart entry requires:

- `name`: The name of the chart
- `repo`: The URL of the Helm repository
- `version`: The version of the chart to include

Optionally, you can specify:

- `valuesFile`: A path to a values file to include with the chart

```yaml
charts:
  - name: redis
    repo: https://charts.bitnami.com/bitnami
    version: 17.11.7
    valuesFile: redis-values.yaml  # Optional
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
```

## Validating a Manifest

Before building a bundle, you can validate your manifest file using the `init` command:

```bash
capsailer init --manifest manifest.yaml
```

This will check that the manifest is properly formatted and that all required fields are present. 