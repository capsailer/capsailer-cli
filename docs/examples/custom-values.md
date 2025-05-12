# Custom Values Example

This example demonstrates how to use custom values files with Capsailer to customize your Helm chart deployments in air-gapped environments.

## Why Use Custom Values?

Custom values files allow you to:

- Configure application-specific settings
- Adapt deployments to different environments
- Override default chart values
- Customize resource allocations
- Configure persistence, security settings, and more

## Scenario: Deploying PostgreSQL with Custom Configuration

In this example, we'll deploy PostgreSQL with custom configuration for an air-gapped environment.

## Step 1: Create Custom Values Files (Connected Environment)

Create a custom values file for PostgreSQL:

```bash
cat > postgresql-values.yaml << EOF
global:
  postgresql:
    auth:
      username: myapp
      password: mypassword
      database: myappdb
  
primary:
  persistence:
    size: 10Gi
  
  resources:
    requests:
      memory: 256Mi
      cpu: 250m
    limits:
      memory: 1Gi
      cpu: 500m

readReplicas:
  replicaCount: 1
EOF
```

## Step 2: Create the Manifest (Connected Environment)

Create a manifest that includes the PostgreSQL chart and references the values file:

```bash
cat > manifest.yaml << EOF
images:
  - registry:2
  - bitnami/postgresql:15.4.0

charts:
  - name: postgresql
    repo: https://charts.bitnami.com/bitnami
    version: 12.5.7
    valuesFile: postgresql-values.yaml
EOF
```

## Step 3: Build the Bundle (Connected Environment)

Build a bundle that includes both the chart and the values file:

```bash
capsailer init --manifest manifest.yaml
capsailer build --manifest manifest.yaml --output postgresql-bundle.tar.gz
```

When you include a `valuesFile` in your chart definition, Capsailer automatically:
1. Validates that the file exists
2. Includes the file in the bundle
3. Makes it available during deployment

## Step 4: Transfer to Air-Gapped Environment

Transfer the bundle and the Capsailer binary to your air-gapped environment:

```bash
scp postgresql-bundle.tar.gz capsailer user@airgapped-server:~/
```

## Step 5: Deploy in Air-Gapped Environment

In the air-gapped environment:

```bash
# Deploy registry
./capsailer registry

# Push artifacts
./capsailer push --bundle postgresql-bundle.tar.gz

# Create namespace
kubectl create namespace database

# Deploy PostgreSQL with custom values
./capsailer deploy --chart postgresql --values postgresql-values.yaml --namespace database
```

Note that you need to specify the values file again during deployment. Capsailer will:
1. Look for the chart locally or in ChartMuseum
2. Load the values file
3. Rewrite any image references to use the local registry
4. Deploy the chart with the custom values

## Step 6: Verify the Deployment

Verify that PostgreSQL is running with your custom configuration:

```bash
# Check the deployment
kubectl get all -n database

# Get the PostgreSQL pod name
POSTGRES_POD=$(kubectl get pods -n database -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}')

# Verify the database exists
kubectl exec -it $POSTGRES_POD -n database -- psql -U myapp -d myappdb -c "\l"
```

## Using Multiple Values Files

You can maintain different values files for different environments:

```bash
# Development values
cat > postgresql-dev.yaml << EOF
primary:
  persistence:
    size: 1Gi
  resources:
    requests:
      memory: 128Mi
readReplicas:
  replicaCount: 0
EOF

# Production values
cat > postgresql-prod.yaml << EOF
primary:
  persistence:
    size: 50Gi
  resources:
    requests:
      memory: 1Gi
readReplicas:
  replicaCount: 2
EOF
```

Then include both in your manifest:

```yaml
charts:
  - name: postgresql
    repo: https://charts.bitnami.com/bitnami
    version: 12.5.7
    valuesFile: postgresql-dev.yaml
```

## Values File Rewriting

When you deploy a chart with a values file, Capsailer automatically rewrites image references to use the local registry. For example, if your values file contains:

```yaml
image:
  repository: bitnami/postgresql
  tag: 15.4.0
```

Capsailer will rewrite it to:

```yaml
image:
  repository: registry.capsailer-registry.svc.cluster.local:5000/bitnami/postgresql
  tag: 15.4.0
```

This ensures that the deployment uses the images from your local registry without requiring manual edits to your values files.

## Best Practices for Custom Values

1. **Version Control**: Keep your values files in version control alongside your application code
2. **Documentation**: Document the purpose and usage of each values file
3. **Minimal Overrides**: Only override values that differ from the defaults
4. **Environment Separation**: Use different values files for different environments
5. **Secret Management**: Consider using Kubernetes secrets for sensitive values rather than including them in values files 