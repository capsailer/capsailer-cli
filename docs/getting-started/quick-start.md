# Quick Start

This guide will help you get started with Capsailer by walking through a basic deployment workflow.

## Prerequisites

Before you begin, make sure you have:

- Capsailer [installed](installation.md) in both your connected and air-gapped environments
- A Kubernetes cluster in your air-gapped environment
- kubectl configured to access your cluster

## Step 1: Create a Manifest File

In your connected environment, create a simple manifest file that defines what you want to deploy:

```bash
cat > manifest.yaml << EOF
images:
  - nginx:1.25.0
  - registry:2
charts:
  - name: nginx
    repo: https://charts.bitnami.com/bitnami
    version: 15.1.4
EOF
```

This manifest includes:
- The Nginx container image
- The registry image (needed for air-gapped deployments)
- The Nginx Helm chart from the Bitnami repository

## Step 2: Build a Bundle

Still in your connected environment, build a bundle from your manifest:

```bash
# Initialize and validate the manifest
capsailer init --manifest manifest.yaml

# Build the bundle
capsailer build --manifest manifest.yaml --output my-bundle.tar.gz
```

This will:
1. Download the specified container images
2. Download the specified Helm charts
3. Package everything into a single archive file

## Step 3: Transfer the Bundle

Transfer the bundle and the Capsailer binary to your air-gapped environment using your organization's approved method:

```bash
# Example using secure copy (if direct transfer is possible)
scp my-bundle.tar.gz capsailer user@airgapped-server:~/
```

## Step 4: Deploy in the Air-Gapped Environment

In your air-gapped environment:

```bash
# Deploy a registry
./capsailer registry

# Push artifacts from the bundle to the registry
./capsailer push --bundle my-bundle.tar.gz

# Deploy the Nginx application
./capsailer deploy --chart nginx
```

## Step 5: Verify the Deployment

Verify that your application is running:

```bash
# Check the pods
kubectl get pods

# Check the services
kubectl get services

# Access the application (using port-forwarding)
kubectl port-forward svc/nginx 8080:80
```

In another terminal, you can access the application:

```bash
curl http://localhost:8080
```

## What's Happening Behind the Scenes

Let's break down what Capsailer is doing:

1. **Registry Deployment**:
   - Deploys a Docker registry and ChartMuseum in your Kubernetes cluster
   - Sets up persistent storage for both services

2. **Artifact Pushing**:
   - Extracts images and charts from the bundle
   - Pushes images to the Docker registry
   - Publishes charts to ChartMuseum

3. **Application Deployment**:
   - Searches for the chart locally or in ChartMuseum
   - Rewrites image references to use the local registry
   - Deploys the application using Helm

## Next Steps

Now that you've completed the quick start, you can:

1. Learn more about the [manifest file format](manifest.md)
2. Explore the [command reference](../commands/index.md)
3. Try the [complete workflow example](../examples/workflow.md)
4. Learn about [air-gapped environments](../concepts/air-gapped.md) 