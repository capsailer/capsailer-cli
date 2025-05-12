# Deploy Command

The `deploy` command installs Helm charts in an air-gapped Kubernetes environment.

## Usage

```bash
capsailer deploy --chart CHART_NAME [flags]
```

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--chart` | Name of the chart to deploy (required) | |
| `--values` | Path to values file for the chart | |
| `--namespace` | Kubernetes namespace to deploy to | `default` |
| `--kubeconfig` | Path to kubeconfig file | `~/.kube/config` |
| `--registry-namespace` | Namespace where registry is deployed | `capsailer-registry` |

## Description

The `deploy` command performs the following steps:

1. **Chart Discovery** - Searches for the specified chart in:
   - Local `charts/` directory
   - ChartMuseum repository in the specified namespace

2. **Values Configuration** - Loads and processes the values file if provided

3. **Image Rewriting** - Automatically rewrites image references in values to use the local registry

4. **Chart Installation** - Installs the chart using Helm

## Chart Discovery Process

The deploy command implements an intelligent chart discovery mechanism:

1. First, it looks for the chart in the local `charts/` directory using these patterns:
   - `charts/CHART_NAME`
   - `charts/CHART_NAME.tgz`
   - `charts/CHART_NAME-*.tgz` (matching by prefix)

2. If not found locally, it checks if ChartMuseum is available in the specified namespace:
   - Sets up port-forwarding to access ChartMuseum
   - Queries the ChartMuseum API to check if the chart exists
   - Downloads the chart from ChartMuseum if found

## Examples

```bash
# Deploy a chart
capsailer deploy --chart nginx

# Deploy with custom values
capsailer deploy --chart nginx --values my-values.yaml --namespace web
```

## Workflow Integration

The `deploy` command is typically used after setting up the registry and pushing artifacts:

```bash
# Set up registry and ChartMuseum
capsailer registry --namespace my-registry

# Push artifacts from a bundle
capsailer push --bundle capsailer-bundle.tar.gz --namespace my-registry

# Deploy a chart
capsailer deploy --chart nginx --registry-namespace my-registry
```

## Troubleshooting

If the deploy command fails to find your chart, check the following:

- Verify that the chart name matches exactly what was included in your bundle
- Ensure the registry and ChartMuseum are running in the specified namespace
- Check that you've pushed the artifacts using the `push` command
- Verify network connectivity to the ChartMuseum service within the cluster

You can manually check if your chart is available in ChartMuseum by port-forwarding and querying the API:

```bash
kubectl port-forward -n capsailer-registry svc/chartmuseum 8080:8080
curl http://localhost:8080/api/charts
``` 