# Complete Workflow Example

This example demonstrates the complete workflow for deploying an application in an air-gapped environment using Capsailer.

## Prerequisites

- Go 1.20 or later installed in the connected environment
- Kubernetes cluster in the air-gapped environment
- kubectl configured to access the air-gapped cluster

## Step 1: Create a Manifest (Connected Environment)

First, create a manifest file that defines the images and charts you want to include in your bundle:

```bash
cat > manifest.yaml << EOF
images:
  - nginx:1.25.0
  - redis:7.0.14
  - bitnami/postgresql:15.4.0
  # Include the registry image to ensure it's available in the air-gapped environment
  - registry:2

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
EOF
```

Create a custom values file for Redis:

```bash
cat > redis-values.yaml << EOF
architecture: standalone
auth:
  enabled: false
master:
  persistence:
    enabled: false
EOF
```

## Step 2: Initialize and Validate the Manifest (Connected Environment)

Validate the manifest to ensure it's correctly formatted:

```bash
capsailer init --manifest manifest.yaml
```

This will output a summary of the images and charts that will be included in the bundle.

## Step 3: Build the Bundle (Connected Environment)

Build a bundle containing all the specified images and charts:

```bash
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz
```

This will:
1. Download all specified container images
2. Download all specified Helm charts
3. Package everything into a single archive file

## Step 4: Transfer the Bundle to the Air-Gapped Environment

Transfer the bundle file to the air-gapped environment using your organization's approved method:

```bash
# Example using secure copy (if direct transfer is possible)
scp capsailer-bundle.tar.gz user@airgapped-server:~/
```

Also transfer the Capsailer binary if it's not already available in the air-gapped environment:

```bash
scp capsailer user@airgapped-server:~/
```

## Step 5: Deploy Registry Infrastructure (Air-Gapped Environment)

In the air-gapped environment, deploy a registry and ChartMuseum:

```bash
# Create a namespace for the registry
kubectl create namespace capsailer-registry

# Deploy the registry
./capsailer registry --namespace capsailer-registry
```

This will:
1. Create a namespace for the registry
2. Deploy a Docker registry and ChartMuseum
3. Set up persistent storage for both services

## Step 6: Push Artifacts to the Registry (Air-Gapped Environment)

Push all container images and Helm charts from the bundle to the registry:

```bash
./capsailer push --bundle capsailer-bundle.tar.gz --namespace capsailer-registry
```

This will:
1. Extract images and charts from the bundle
2. Push images to the Docker registry
3. Publish charts to ChartMuseum

## Step 7: Deploy Applications (Air-Gapped Environment)

Now you can deploy applications using the charts in the registry:

```bash
# Create namespaces for your applications
kubectl create namespace web
kubectl create namespace database

# Deploy Nginx
./capsailer deploy --chart nginx --registry-namespace capsailer-registry --namespace web

# Deploy Redis with custom values
./capsailer deploy --chart redis --values redis-values.yaml --registry-namespace capsailer-registry --namespace database

# Deploy PostgreSQL
./capsailer deploy --chart postgresql --registry-namespace capsailer-registry --namespace database
```

Each deployment will:
1. Find the chart (either locally or in ChartMuseum)
2. Rewrite image references to use the local registry
3. Install the chart in the specified namespace

## Step 8: Verify the Deployments

Verify that the applications are running correctly:

```bash
# Check Nginx deployment
kubectl get all -n web

# Check Redis deployment
kubectl get all -n database
```

## Complete Script

Here's the complete script for the entire workflow:

```bash
# CONNECTED ENVIRONMENT

# Create manifest
cat > manifest.yaml << EOF
images:
  - nginx:1.25.0
  - redis:7.0.14
  - bitnami/postgresql:15.4.0
  - registry:2
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
EOF

# Create values file
cat > redis-values.yaml << EOF
architecture: standalone
auth:
  enabled: false
master:
  persistence:
    enabled: false
EOF

# Initialize and build
capsailer init --manifest manifest.yaml
capsailer build --manifest manifest.yaml --output capsailer-bundle.tar.gz

# Transfer files (example)
scp capsailer-bundle.tar.gz user@airgapped-server:~/
scp capsailer redis-values.yaml user@airgapped-server:~/

# AIR-GAPPED ENVIRONMENT

# Deploy registry
kubectl create namespace capsailer-registry
./capsailer registry --namespace capsailer-registry

# Push artifacts
./capsailer push --bundle capsailer-bundle.tar.gz --namespace capsailer-registry

# Deploy applications
kubectl create namespace web
kubectl create namespace database
./capsailer deploy --chart nginx --registry-namespace capsailer-registry --namespace web
./capsailer deploy --chart redis --values redis-values.yaml --registry-namespace capsailer-registry --namespace database
./capsailer deploy --chart postgresql --registry-namespace capsailer-registry --namespace database

# Verify deployments
kubectl get all -n web
kubectl get all -n database
```

This workflow demonstrates the complete process of using Capsailer to deploy applications in an air-gapped environment, from manifest creation to verification of the deployed applications. 