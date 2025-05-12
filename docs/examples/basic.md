# Basic Usage

This page provides a simple example of using Capsailer to deploy a basic web application in an air-gapped environment.

## Scenario

In this example, we'll deploy a simple Nginx web server in an air-gapped Kubernetes cluster.

## Connected Environment Steps

### 1. Install Capsailer

First, install Capsailer in your connected environment:

```bash
# Clone the repository
git clone https://github.com/jlnhnng/capsailer.git
cd capsailer

# Build the binary
go build -o capsailer cmd/capsailer/main.go cmd/capsailer/commands.go

# Add to your PATH
sudo mv capsailer /usr/local/bin/
```

### 2. Create a Manifest

Create a simple manifest file that includes the Nginx image and chart:

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

### 3. Build the Bundle

Build a bundle containing the Nginx image and chart:

```bash
capsailer build --manifest manifest.yaml --output nginx-bundle.tar.gz
```

## Air-Gapped Environment Steps

### 1. Transfer the Bundle

Transfer the bundle and the Capsailer binary to your air-gapped environment:

```bash
# Example using secure copy (if direct transfer is possible)
scp nginx-bundle.tar.gz capsailer user@airgapped-server:~/
```

### 2. Deploy the Registry

In the air-gapped environment, deploy a registry:

```bash
./capsailer registry
```

### 3. Push the Bundle Contents

Push the contents of the bundle to the registry:

```bash
./capsailer push --bundle nginx-bundle.tar.gz
```

### 4. Deploy Nginx

Deploy the Nginx chart:

```bash
./capsailer deploy --chart nginx
```

### 5. Verify the Deployment

Verify that Nginx is running:

```bash
kubectl get pods
kubectl get services
```

Access the Nginx service:

```bash
# Port-forward the Nginx service
kubectl port-forward svc/nginx 8080:80

# In another terminal
curl http://localhost:8080
```

## Explanation

This basic example demonstrates the core workflow of Capsailer:

1. **Define Requirements**: Create a manifest specifying what you need
2. **Build Bundle**: Package everything into a portable archive
3. **Transfer Bundle**: Move the bundle to the air-gapped environment
4. **Deploy Infrastructure**: Set up a registry in the air-gapped environment
5. **Push Artifacts**: Upload images and charts to the registry
6. **Deploy Application**: Deploy the application using the local artifacts

The key feature demonstrated here is the ability to deploy an application in an environment without internet access, using only the artifacts that were bundled in the connected environment.

## Next Steps

After mastering this basic example, you can:

- Try deploying more complex applications
- Use custom values files to customize deployments
- Include multiple applications in a single bundle
- Explore the [Complete Workflow Example](workflow.md) for a more comprehensive scenario 