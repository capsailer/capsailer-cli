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
	PersistentPV   bool
	KubeconfigPath string
}

// DefaultRegistryOptions returns default registry options
func DefaultRegistryOptions() RegistryOptions {
	return RegistryOptions{
		Namespace:      "capsailer-registry",
		RegistryImage:  "registry:2",
		PersistentPV:   true,
		KubeconfigPath: "",
	}
}

// SetupRegistry sets up a Docker registry in a Kubernetes cluster
func SetupRegistry(opts RegistryOptions) (string, error) {
	// Create YAML file for the registry
	manifestPath, err := createRegistryManifest(opts)
	if err != nil {
		return "", fmt.Errorf("failed to create registry manifest: %w", err)
	}

	// Apply manifest using kubectl
	kubectlArgs := []string{"apply", "-f", manifestPath}
	if opts.KubeconfigPath != "" {
		kubectlArgs = append(kubectlArgs, "--kubeconfig", opts.KubeconfigPath)
	}

	fmt.Println("Deploying registry to Kubernetes cluster...")
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

	// Return the registry URL
	registryURL := fmt.Sprintf("registry.%s.svc.cluster.local:5000", opts.Namespace)
	return registryURL, nil
}

// createRegistryManifest creates a YAML manifest for the registry
func createRegistryManifest(opts RegistryOptions) (string, error) {
	// Create a temporary file for the manifest
	manifestPath := filepath.Join(os.TempDir(), "capsailer-registry.yaml")

	// Define the manifest content
	var volumeSection string
	var volumeMountSection string
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
`, opts.Namespace)

		volumeMountSection = `
      volumes:
        - name: registry-data
          persistentVolumeClaim:
            claimName: registry-data`
	} else {
		volumeMountSection = `
      volumes:
        - name: registry-data
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
`, opts.Namespace, volumeSection, opts.Namespace, opts.RegistryImage, volumeMountSection, opts.Namespace)

	// Write the manifest to the file
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		return "", fmt.Errorf("failed to write registry manifest: %w", err)
	}

	fmt.Printf("Created registry manifest at %s\n", manifestPath)
	return manifestPath, nil
} 