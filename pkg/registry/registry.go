package registry

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RegistryOptions defines options for the registry
type RegistryOptions struct {
	Namespace      string
	RegistryImage  string
	ChartMuseumImage string
	PersistentPV   bool
	KubeconfigPath string
}

// DefaultRegistryOptions returns default registry options
func DefaultRegistryOptions() RegistryOptions {
	return RegistryOptions{
		Namespace:        "capsailer-registry",
		RegistryImage:    "registry:2",
		ChartMuseumImage: "ghcr.io/helm/chartmuseum:v0.15.0",
		PersistentPV:     true,
		KubeconfigPath:   "",
	}
}

// SetupRegistry sets up a Docker registry and ChartMuseum in a Kubernetes cluster
func SetupRegistry(opts RegistryOptions) (string, error) {
	// Create YAML file for the registry and ChartMuseum
	manifestPath, err := createRegistryManifest(opts)
	if err != nil {
		return "", fmt.Errorf("failed to create registry manifest: %w", err)
	}

	// Apply manifest using kubectl
	kubectlArgs := []string{"apply", "-f", manifestPath}
	if opts.KubeconfigPath != "" {
		kubectlArgs = append(kubectlArgs, "--kubeconfig", opts.KubeconfigPath)
	}

	fmt.Println("Deploying registry and ChartMuseum to Kubernetes cluster...")
	cmd := exec.Command("kubectl", kubectlArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to apply registry manifest: %w", err)
	}

	// Wait for registry to be ready
	fmt.Println("Waiting for registry deployment to be ready...")
	waitArgs := []string{"rollout", "status", "deployment/registry", "-n", opts.Namespace}
	if opts.KubeconfigPath != "" {
		waitArgs = append(waitArgs, "--kubeconfig", opts.KubeconfigPath)
	}

	waitCmd := exec.Command("kubectl", waitArgs...)
	waitCmd.Stdout = os.Stdout
	waitCmd.Stderr = os.Stderr
	if err := waitCmd.Run(); err != nil {
		return "", fmt.Errorf("failed waiting for registry: %w", err)
	}

	// Wait for ChartMuseum to be ready
	fmt.Println("Waiting for ChartMuseum deployment to be ready...")
	waitChartArgs := []string{"rollout", "status", "deployment/chartmuseum", "-n", opts.Namespace}
	if opts.KubeconfigPath != "" {
		waitChartArgs = append(waitChartArgs, "--kubeconfig", opts.KubeconfigPath)
	}

	waitChartCmd := exec.Command("kubectl", waitChartArgs...)
	waitChartCmd.Stdout = os.Stdout
	waitChartCmd.Stderr = os.Stderr
	if err := waitChartCmd.Run(); err != nil {
		return "", fmt.Errorf("failed waiting for ChartMuseum: %w", err)
	}

	// Return the registry URL
	registryURL := fmt.Sprintf("registry.%s.svc.cluster.local:5000", opts.Namespace)
	return registryURL, nil
}

// createRegistryManifest creates a YAML manifest for the registry and ChartMuseum
func createRegistryManifest(opts RegistryOptions) (string, error) {
	// Create a temporary file for the manifest
	manifestPath := filepath.Join(os.TempDir(), "capsailer-registry.yaml")

	// Define the manifest content
	var volumeSection string
	var volumeMountSection string
	var chartVolumeMountSection string
	
	if opts.PersistentPV {
		volumeSection = fmt.Sprintf(`
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: registry-data
  namespace: %s
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: chartmuseum-data
  namespace: %s
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
`, opts.Namespace, opts.Namespace)

		volumeMountSection = `
      volumes:
        - name: registry-data
          persistentVolumeClaim:
            claimName: registry-data`
            
		chartVolumeMountSection = `
      volumes:
        - name: chartmuseum-data
          persistentVolumeClaim:
            claimName: chartmuseum-data`
	} else {
		volumeMountSection = `
      volumes:
        - name: registry-data
          emptyDir: {}`
		chartVolumeMountSection = `
      volumes:
        - name: chartmuseum-data
          emptyDir: {}`
	}

	manifest := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
%s
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: registry
  namespace: %s
spec:
  selector:
    matchLabels:
      app: registry
  replicas: 1
  template:
    metadata:
      labels:
        app: registry
    spec:
      containers:
        - name: registry
          image: %s
          ports:
            - containerPort: 5000
          volumeMounts:
            - name: registry-data
              mountPath: /var/lib/registry
          env:
            - name: REGISTRY_STORAGE_DELETE_ENABLED
              value: "true"%s
---
apiVersion: v1
kind: Service
metadata:
  name: registry
  namespace: %s
spec:
  selector:
    app: registry
  ports:
    - port: 5000
      targetPort: 5000
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chartmuseum
  namespace: %s
spec:
  selector:
    matchLabels:
      app: chartmuseum
  replicas: 1
  template:
    metadata:
      labels:
        app: chartmuseum
    spec:
      containers:
        - name: chartmuseum
          image: %s
          ports:
            - containerPort: 8080
          volumeMounts:
            - name: chartmuseum-data
              mountPath: /storage
          env:
            - name: DEBUG
              value: "false"
            - name: STORAGE
              value: "local"
            - name: STORAGE_LOCAL_ROOTDIR
              value: "/storage"
            - name: ALLOW_OVERWRITE
              value: "true"
            - name: DISABLE_API
              value: "false"%s
---
apiVersion: v1
kind: Service
metadata:
  name: chartmuseum
  namespace: %s
spec:
  selector:
    app: chartmuseum
  ports:
    - port: 8080
      targetPort: 8080
  type: ClusterIP
`, opts.Namespace, volumeSection, opts.Namespace, opts.RegistryImage, volumeMountSection, opts.Namespace, opts.Namespace, opts.ChartMuseumImage, chartVolumeMountSection, opts.Namespace)

	// Write the manifest to the file
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		return "", fmt.Errorf("failed to write registry manifest: %w", err)
	}

	fmt.Printf("Created registry manifest at %s\n", manifestPath)
	return manifestPath, nil
}
