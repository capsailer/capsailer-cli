package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/capsailer/capsailer-cli/pkg/registry"
	"github.com/spf13/cobra"
)

// runRegistry handles the registry command
func runRegistry(namespace, image string, persistent bool, kubeconfigPath string) error {
	fmt.Println("Deploying a standalone Docker registry")

	// Air-gapped environment handling
	fmt.Println("\nAir-gapped environment detection:")
	isAirGapped := detectAirGapped()

	if isAirGapped {
		fmt.Println("Air-gapped environment detected.")
		fmt.Println("Checking for registry image in local bundle...")

		// Check if we have the registry image in a local bundle
		if _, err := os.Stat("images/registry_2.tar"); os.IsNotExist(err) {
			fmt.Println("Registry image not found in local bundle.")
			fmt.Println("Options for air-gapped registry deployment:")
			fmt.Println("1. Pre-load the registry image on your cluster nodes")
			fmt.Println("2. Transfer the registry image manually to your cluster")
			fmt.Println("3. Run 'capsailer build' with a manifest that includes 'registry:2'")
			fmt.Println("   then unpack that bundle first")

			// Ask if they want to proceed
			fmt.Println("\nThe deployment might fail if the registry image is not available.")
			fmt.Print("Do you want to proceed with deployment? (y/n): ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				return fmt.Errorf("deployment cancelled by user")
			}
		} else {
			fmt.Println("Registry image found in local bundle. It will be used for deployment.")
			// Here we would load the image into the cluster nodes first
			fmt.Println("Loading registry image from local bundle...")
			err := loadImageToCluster("images/registry_2.tar", kubeconfigPath)
			if err != nil {
				fmt.Printf("Warning: Failed to load image: %v\n", err)
				fmt.Println("Will attempt to continue deployment assuming the image is available in the cluster.")
			}
		}
	} else {
		fmt.Println("Connected environment detected. Registry image will be pulled from Docker Hub.")
	}

	// Create registry options
	opts := registry.RegistryOptions{
		Namespace:      namespace,
		RegistryImage:  image,
		PersistentPV:   persistent,
		KubeconfigPath: kubeconfigPath,
	}

	// Setup the registry
	registryURL, err := registry.SetupRegistry(opts)
	if err != nil {
		return fmt.Errorf("failed to setup registry: %w", err)
	}

	fmt.Printf("Registry deployed successfully at: %s\n", registryURL)
	fmt.Println("\nYou can use this registry for your air-gapped deployments.")
	fmt.Println("To push images to this registry:")
	fmt.Printf("  docker tag myimage:tag %s/myimage:tag\n", registryURL)
	fmt.Printf("  docker push %s/myimage:tag\n", registryURL)

	return nil
}

// detectAirGapped attempts to determine if we're in an air-gapped environment
func detectAirGapped() bool {
	// Try to access a well-known internet endpoint
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	_, err := client.Get("https://registry-1.docker.io/v2/")

	// If there's an error, assume we're air-gapped
	return err != nil
}

// loadImageToCluster loads a container image from a local tar file to the cluster
func loadImageToCluster(imageTarPath string, kubeconfigPath string) error {
	// First load the image into the local docker daemon
	fmt.Printf("Loading image from %s into local docker daemon...\n", imageTarPath)
	loadCmd := exec.Command("docker", "load", "-i", imageTarPath)
	loadCmd.Stdout = os.Stdout
	loadCmd.Stderr = os.Stderr
	if err := loadCmd.Run(); err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	// Get nodes in the cluster
	var kubectlArgs []string
	kubectlArgs = append(kubectlArgs, "get", "nodes", "-o", "jsonpath='{.items[*].metadata.name}'")
	if kubeconfigPath != "" {
		kubectlArgs = append(kubectlArgs, "--kubeconfig", kubeconfigPath)
	}

	cmd := exec.Command("kubectl", kubectlArgs...)
	nodesOutput, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get cluster nodes: %w", err)
	}

	// Parse node names (simple space-separated string)
	nodeNames := strings.Split(strings.Trim(string(nodesOutput), "'"), " ")

	// For real clusters we would distribute the image to each node
	// In a single-node or minikube cluster, the local docker load is sufficient
	fmt.Printf("Found %d nodes in cluster\n", len(nodeNames))

	if len(nodeNames) > 1 {
		fmt.Println("Note: For multi-node clusters, you may need additional steps to ensure")
		fmt.Println("the registry image is available on all nodes. Options include:")
		fmt.Println("1. Use your cluster's image distribution mechanism if available")
		fmt.Println("2. Manually load the image on each node")
		fmt.Println("3. Configure nodes to pull from a local registry (if available)")
	}

	return nil
}

// runPush handles the push command
func runPush(image, bundlePath, namespace, kubeconfigPath string, externalRegistry, username, password string) error {
	var registryURL string

	if externalRegistry != "" {
		// Use the external registry URL
		fmt.Printf("Using external registry: %s\n", externalRegistry)
		registryURL = externalRegistry

		// If credentials are provided, attempt to log in
		if username != "" {
			fmt.Println("Authenticating with registry...")
			if err := loginToRegistry(externalRegistry, username, password); err != nil {
				return fmt.Errorf("failed to authenticate with registry: %w", err)
			}
		}
	} else {
		// Get registry URL from the Kubernetes service
		var kubectlArgs []string
		kubectlArgs = append(kubectlArgs, "get", "service", "-n", namespace, "registry", "-o", "jsonpath='{.spec.clusterIP}'")
		if kubeconfigPath != "" {
			kubectlArgs = append(kubectlArgs, "--kubeconfig", kubeconfigPath)
		}

		fmt.Println("Finding registry service...")
		cmd := exec.Command("kubectl", kubectlArgs...)
		registryIPOutput, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get registry service: %w", err)
		}

		registryIP := strings.Trim(string(registryIPOutput), "'")
		if registryIP == "" {
			return fmt.Errorf("registry service not found in namespace %s", namespace)
		}

		registryURL = fmt.Sprintf("%s:5000", registryIP)
		fmt.Printf("Found registry at %s\n", registryURL)

		// Set up port-forwarding to the registry to make it accessible from the CLI
		// This is necessary since we're pushing directly instead of using an in-cluster tool
		fmt.Printf("Setting up port-forwarding to registry in namespace %s...\n", namespace)
		forwardProc, err := startPortForward(namespace, "registry", 5000, kubeconfigPath)
		if err != nil {
			fmt.Printf("Warning: Failed to set up port forwarding: %v\n", err)
			fmt.Println("Will attempt to push directly to the registry ClusterIP (may fail if not reachable)")
		} else {
			defer stopPortForward(forwardProc)
			// Use localhost URL since we've set up port forwarding
			registryURL = "localhost:5000"
			fmt.Printf("Port forwarding established, using registry at %s\n", registryURL)
		}
	}

	// Handle different push modes
	if bundlePath != "" {
		// Push all artifacts from a bundle
		if err := pushImagesFromBundle(bundlePath, registryURL, namespace, kubeconfigPath); err != nil {
			return fmt.Errorf("failed to push images: %w", err)
		}

		// Push charts if they exist and we're not using an external registry
		// (since chart publishing requires ChartMuseum)
		if externalRegistry == "" {
			if err := publishChartsFromBundle(bundlePath, namespace, kubeconfigPath); err != nil {
				return fmt.Errorf("failed to publish charts: %w", err)
			}
		} else {
			fmt.Println("Skipping chart publishing for external registry.")
			fmt.Println("Charts can only be published to the built-in ChartMuseum repository.")
		}

		return nil
	} else if image != "" {
		// Push a single image
		return pushSingleImage(image, registryURL)
	}

	return fmt.Errorf("either --image or --bundle must be specified")
}

// loginToRegistry attempts to authenticate with an external registry
func loginToRegistry(registry, username, password string) error {
	// Try to authenticate using Docker CLI if available
	if checkCommandAvailable("docker") {
		fmt.Println("Using Docker CLI for authentication...")
		loginCmd := exec.Command("docker", "login", registry, "-u", username, "--password-stdin")
		loginCmd.Stdin = strings.NewReader(password)
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr

		if err := loginCmd.Run(); err != nil {
			return fmt.Errorf("docker login failed: %w", err)
		}
		fmt.Println("Successfully authenticated with Docker CLI")
		return nil
	}

	// If Docker is not available, we'll just use the credentials when pushing
	// We can't easily verify the credentials without making an actual API call
	fmt.Println("Docker CLI not available, will use credentials directly when pushing images")
	return nil
}

// pushSingleImage pushes a single image to the registry
func pushSingleImage(image, registryURL string) error {
	fmt.Printf("Pushing single image %s to registry\n", image)

	// Check if image exists locally
	fmt.Printf("Checking if image %s exists locally...\n", image)
	inspectCmd := exec.Command("docker", "image", "inspect", image)
	if err := inspectCmd.Run(); err != nil {
		return fmt.Errorf("image %s not found locally: %w", image, err)
	}

	// Create a tagged version for the registry
	parts := strings.Split(image, "/")
	var imageName string
	if len(parts) > 1 {
		imageName = parts[len(parts)-1]
	} else {
		imageName = image
	}

	targetImage := fmt.Sprintf("%s/%s", registryURL, imageName)
	fmt.Printf("Tagging image as %s\n", targetImage)

	tagCmd := exec.Command("docker", "tag", image, targetImage)
	tagCmd.Stdout = os.Stdout
	tagCmd.Stderr = os.Stderr
	if err := tagCmd.Run(); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}

	// Push the image to the registry
	fmt.Printf("Pushing image to registry at %s\n", registryURL)
	pushCmd := exec.Command("docker", "push", targetImage)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	fmt.Printf("Successfully pushed %s to registry\n", image)
	fmt.Printf("Image is now available as: %s\n", targetImage)

	return nil
}

// pushImagesFromBundle pushes all images from a bundle to a registry
func pushImagesFromBundle(bundlePath, registryURL, namespace, kubeconfigPath string) error {
	fmt.Printf("Pushing all images from bundle %s to registry\n", bundlePath)

	// First, check if the bundle exists
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		return fmt.Errorf("bundle file not found: %s", bundlePath)
	}

	// Create a temporary directory for unpacking
	tempDir, err := os.MkdirTemp("", "capsailer-bundle-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing temp directory: %v\n", err)
		}
	}()

	// Determine the images directory path
	var imagesDir string
	if filepath.Ext(bundlePath) == ".tar" || filepath.Ext(bundlePath) == ".gz" {
		fmt.Printf("Extracting bundle to %s...\n", tempDir)

		// Extract the bundle
		extractCmd := exec.Command("tar", "-xf", bundlePath, "-C", tempDir)
		extractCmd.Stdout = os.Stdout
		extractCmd.Stderr = os.Stderr
		if err := extractCmd.Run(); err != nil {
			return fmt.Errorf("failed to extract bundle: %w", err)
		}

		imagesDir = filepath.Join(tempDir, "images")
	} else {
		// Assume bundlePath is a directory that might contain an images directory
		potentialImagesDir := filepath.Join(bundlePath, "images")
		if _, err := os.Stat(potentialImagesDir); !os.IsNotExist(err) {
			imagesDir = potentialImagesDir
		} else {
			// If no images subdirectory, assume the specified path is the images directory
			imagesDir = bundlePath
		}
	}

	// Check if images directory exists
	if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
		return fmt.Errorf("images directory not found in bundle: %s", imagesDir)
	}

	// Get list of image tars in the images directory
	imageTars, err := filepath.Glob(filepath.Join(imagesDir, "*.tar"))
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	if len(imageTars) == 0 {
		return fmt.Errorf("no image tars found in %s", imagesDir)
	}

	fmt.Printf("Found %d image tars to push\n", len(imageTars))

	// Process each image tar
	for _, imageTar := range imageTars {
		imageName := filepath.Base(imageTar)
		imageName = strings.TrimSuffix(imageName, ".tar")

		fmt.Printf("Processing image %s\n", imageName)

		// Extract original image name from tarball name
		// Convert underscores back to slashes and colons
		repoPath := imageName
		if strings.Contains(repoPath, "_") {
			// Last underscore is likely separating the tag
			lastUnderscore := strings.LastIndex(repoPath, "_")
			if lastUnderscore != -1 {
				repoPath = strings.ReplaceAll(repoPath[:lastUnderscore], "_", "/") + ":" + repoPath[lastUnderscore+1:]
			}
		}

		// Target reference for the image in the registry
		targetRef := fmt.Sprintf("%s/%s", registryURL, repoPath)
		fmt.Printf("Pushing image to %s\n", targetRef)

		// Load the image from tar file
		fmt.Printf("Loading image from %s...\n", imageTar)

		// Direct implementation using go-containerregistry
		if err := pushImageToRegistry(imageTar, targetRef); err != nil {
			fmt.Printf("Warning: Failed to push image: %v\n", err)

			// Try using Docker if available
			if checkCommandAvailable("docker") {
				fmt.Println("Attempting to push with Docker as fallback...")
				if err := pushWithDocker(imageTar, targetRef); err != nil {
					fmt.Printf("Warning: Docker fallback also failed: %v\n", err)
					fmt.Printf("Manual steps to push this image:\n")
					fmt.Printf("  1. Load the image: docker load -i %s\n", imageTar)
					fmt.Printf("  2. Tag the image: docker tag %s %s\n", repoPath, targetRef)
					fmt.Printf("  3. Push the image: docker push %s\n", targetRef)
				} else {
					fmt.Printf("Successfully pushed image using Docker: %s\n", targetRef)
				}
			} else {
				fmt.Printf("Manual steps to push this image:\n")
				fmt.Printf("  1. Load the image: docker load -i %s\n", imageTar)
				fmt.Printf("  2. Tag the image: docker tag %s %s\n", repoPath, targetRef)
				fmt.Printf("  3. Push the image: docker push %s\n", targetRef)
			}
		} else {
			fmt.Printf("Successfully pushed image: %s\n", targetRef)
		}
	}

	fmt.Printf("All images from bundle have been processed.\n")
	return nil
}

// checkCommandAvailable checks if a command is available in the PATH
func checkCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// pushWithDocker uses Docker as a fallback method to push images
func pushWithDocker(imageTarPath, targetRef string) error {
	// Load the image
	loadCmd := exec.Command("docker", "load", "-i", imageTarPath)
	loadOutput, err := loadCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	// Parse the image name from the output
	loadedName := parseDockerLoadOutput(string(loadOutput), targetRef)

	// Tag the image
	tagCmd := exec.Command("docker", "tag", loadedName, targetRef)
	if err := tagCmd.Run(); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}

	// Push the image
	pushCmd := exec.Command("docker", "push", targetRef)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	return nil
}

// parseDockerLoadOutput parses the output of docker load to get the image name
func parseDockerLoadOutput(output, defaultName string) string {
	if strings.Contains(output, "Loaded image:") {
		parts := strings.Split(output, "Loaded image:")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	} else if strings.Contains(output, "Loaded image ID:") {
		// If we got an image ID, we need to use the original name
		fmt.Println("Image loaded by ID, using provided name")
		// Extract just the image name without the registry
		parts := strings.Split(defaultName, "/")
		if len(parts) > 1 {
			return strings.Join(parts[1:], "/")
		}
	}

	// Default to the original name
	return defaultName
}

// pushImageToRegistry pushes an image from a tar file to a registry using go-containerregistry
// This eliminates the dependency on Docker or skopeo
func pushImageToRegistry(imageTarPath, targetRef string) error {
	// Import the image from the tar file
	tag, err := name.NewTag(targetRef)
	if err != nil {
		return fmt.Errorf("invalid target reference: %w", err)
	}

	// Load the image from the tar file
	img, err := tarball.ImageFromPath(imageTarPath, nil)
	if err != nil {
		return fmt.Errorf("failed to load image from tar: %w", err)
	}

	// Push the image to the registry
	// Set up options to allow insecure registries (commonly used in air-gapped environments)
	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Check if we need to use authentication
	// If the registry is not localhost, try to get credentials from Docker config
	var auth authn.Authenticator = authn.Anonymous
	if !strings.HasPrefix(targetRef, "localhost:") {
		// Try to get credentials from Docker config
		auth, err = authn.DefaultKeychain.Resolve(tag.Registry)
		if err != nil {
			fmt.Printf("Warning: Failed to get credentials from Docker config: %v\n", err)
			fmt.Println("Continuing with anonymous authentication")
			auth = authn.Anonymous
		}
	}

	if err := remote.Write(tag, img,
		remote.WithTransport(insecureTransport),
		remote.WithAuth(auth)); err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}

	return nil
}

// publishChartsFromBundle publishes Helm charts from a bundle to a chart repository
func publishChartsFromBundle(bundlePath, namespace, kubeconfigPath string) error {
	fmt.Println("Looking for Helm charts in bundle...")

	// Create a temporary directory for unpacking
	tempDir, err := os.MkdirTemp("", "capsailer-bundle-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing temp directory: %v\n", err)
		}
	}()

	// Determine the charts directory path
	var chartsDir string
	if filepath.Ext(bundlePath) == ".tar" || filepath.Ext(bundlePath) == ".gz" {
		// Bundle is already extracted in pushImagesFromBundle, reuse that temp dir if possible
		// But handle the case if this function is called independently

		// Extract the bundle
		extractCmd := exec.Command("tar", "-xf", bundlePath, "-C", tempDir)
		extractCmd.Stdout = os.Stdout
		extractCmd.Stderr = os.Stderr
		if err := extractCmd.Run(); err != nil {
			return fmt.Errorf("failed to extract bundle: %w", err)
		}

		chartsDir = filepath.Join(tempDir, "charts")
	} else {
		// Assume bundlePath is a directory that might contain a charts directory
		potentialChartsDir := filepath.Join(bundlePath, "charts")
		if _, err := os.Stat(potentialChartsDir); !os.IsNotExist(err) {
			chartsDir = potentialChartsDir
		} else {
			// If no charts subdirectory, assume the specified path is the charts directory
			chartsDir = bundlePath
		}
	}

	// Check if charts directory exists
	if _, err := os.Stat(chartsDir); os.IsNotExist(err) {
		fmt.Println("No charts directory found in bundle, skipping chart publishing")
		return nil
	}

	// Get list of chart tgz files in the charts directory
	chartTgzs, err := filepath.Glob(filepath.Join(chartsDir, "*.tgz"))
	if err != nil {
		return fmt.Errorf("failed to list charts: %w", err)
	}

	if len(chartTgzs) == 0 {
		fmt.Println("No chart packages found in bundle, skipping chart publishing")
		return nil
	}

	fmt.Printf("Found %d charts to publish\n", len(chartTgzs))

	// Set up a Helm chart server in Kubernetes if one doesn't already exist
	if err := setupChartRepository(namespace, kubeconfigPath); err != nil {
		return fmt.Errorf("failed to setup chart repository: %w", err)
	}

	// Get the chart repository URL
	repoURL, err := getChartRepoURL(namespace, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to get chart repository URL: %w", err)
	}

	// Port-forward the chartmuseum service to make it accessible to the CLI
	// This is needed since we're publishing directly from the CLI, not from inside the cluster
	forwardPort, err := startPortForward(namespace, "chartmuseum", 8080, kubeconfigPath)
	if err != nil {
		fmt.Printf("Warning: Could not set up port forwarding to chartmuseum: %v\n", err)
		fmt.Println("Will attempt to publish directly to cluster IP...")
	} else {
		defer stopPortForward(forwardPort)
		// Use localhost for publishing since we have port forwarding
		repoURL = "http://localhost:8080"
	}

	// Publish each chart
	for _, chartTgz := range chartTgzs {
		fmt.Printf("Publishing chart: %s\n", filepath.Base(chartTgz))

		// In a full implementation, we would use a Helm chart repository client
		// to publish the chart. For this example, we'll use a simple HTTP POST
		// to the chartmuseum API.
		if err := publishChartToRepo(chartTgz, repoURL); err != nil {
			return fmt.Errorf("failed to publish chart %s: %w", filepath.Base(chartTgz), err)
		}
	}

	fmt.Printf("All charts have been published to the repository at %s\n", repoURL)
	return nil
}

// startPortForward starts port forwarding to a Kubernetes service
func startPortForward(namespace, serviceName string, port int, kubeconfigPath string) (*os.Process, error) {
	fmt.Printf("Setting up port forwarding to %s in namespace %s...\n", serviceName, namespace)

	var kubectlArgs []string
	kubectlArgs = append(kubectlArgs, "port-forward", "-n", namespace, fmt.Sprintf("svc/%s", serviceName), fmt.Sprintf("%d:%d", port, port))
	if kubeconfigPath != "" {
		kubectlArgs = append(kubectlArgs, "--kubeconfig", kubeconfigPath)
	}

	cmd := exec.Command("kubectl", kubectlArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the command without waiting for it to complete
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start port-forward: %w", err)
	}

	// Give it a moment to establish the connection
	time.Sleep(2 * time.Second)

	return cmd.Process, nil
}

// stopPortForward stops a port forwarding process
func stopPortForward(process *os.Process) {
	if process != nil {
		fmt.Println("Stopping port forwarding...")
		if err := process.Kill(); err != nil {
			fmt.Fprintf(os.Stderr, "Error stopping port forwarding: %v\n", err)
		}
	}
}

// setupChartRepository ensures a Helm chart repository is running
func setupChartRepository(namespace, kubeconfigPath string) error {
	fmt.Println("Setting up Helm chart repository...")

	// See if chartmuseum is already running
	var kubectlArgs []string
	kubectlArgs = append(kubectlArgs, "get", "deployment", "-n", namespace, "chartmuseum", "--ignore-not-found")
	if kubeconfigPath != "" {
		kubectlArgs = append(kubectlArgs, "--kubeconfig", kubeconfigPath)
	}

	cmd := exec.Command("kubectl", kubectlArgs...)
	output, err := cmd.CombinedOutput()
	if err == nil && len(output) > 0 && !strings.Contains(string(output), "No resources found") {
		fmt.Println("Chart repository is already running")
		return nil
	}

	// Create chartmuseum deployment
	chartMuseumManifest, err := createChartMuseumManifest(namespace)
	if err != nil {
		return err
	}

	// Apply the manifest
	applyArgs := []string{"apply", "-f", chartMuseumManifest}
	if kubeconfigPath != "" {
		applyArgs = append(applyArgs, "--kubeconfig", kubeconfigPath)
	}

	fmt.Println("Creating chart repository...")
	applyCmd := exec.Command("kubectl", applyArgs...)
	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr
	if err := applyCmd.Run(); err != nil {
		return fmt.Errorf("failed to create chart repository: %w", err)
	}

	// Wait for deployment to be ready
	waitArgs := []string{"rollout", "status", "deployment/chartmuseum", "-n", namespace}
	if kubeconfigPath != "" {
		waitArgs = append(waitArgs, "--kubeconfig", kubeconfigPath)
	}

	fmt.Println("Waiting for chart repository to be ready...")
	waitCmd := exec.Command("kubectl", waitArgs...)
	waitCmd.Stdout = os.Stdout
	waitCmd.Stderr = os.Stderr
	if err := waitCmd.Run(); err != nil {
		return fmt.Errorf("failed waiting for chart repository: %w", err)
	}

	return nil
}

// createChartMuseumManifest creates a YAML manifest for chartmuseum
func createChartMuseumManifest(namespace string) (string, error) {
	manifest := fmt.Sprintf(`apiVersion: apps/v1
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
          image: chartmuseum/chartmuseum:latest
          ports:
            - containerPort: 8080
          env:
            - name: PORT
              value: "8080"
            - name: DEBUG
              value: "true"
            - name: STORAGE
              value: "local"
            - name: STORAGE_LOCAL_ROOTDIR
              value: "/charts"
            - name: ALLOW_OVERWRITE
              value: "true"
          volumeMounts:
            - name: charts-data
              mountPath: /charts
      volumes:
        - name: charts-data
          emptyDir: {}
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
`, namespace, namespace)

	// Create a temporary file for the manifest
	manifestPath := filepath.Join(os.TempDir(), "capsailer-chartmuseum.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		return "", fmt.Errorf("failed to write chartmuseum manifest: %w", err)
	}

	fmt.Printf("Created chartmuseum manifest at %s\n", manifestPath)
	return manifestPath, nil
}

// getChartRepoURL gets the URL for the chart repository
func getChartRepoURL(namespace, kubeconfigPath string) (string, error) {
	var kubectlArgs []string
	kubectlArgs = append(kubectlArgs, "get", "service", "-n", namespace, "chartmuseum", "-o", "jsonpath='{.spec.clusterIP}'")
	if kubeconfigPath != "" {
		kubectlArgs = append(kubectlArgs, "--kubeconfig", kubeconfigPath)
	}

	cmd := exec.Command("kubectl", kubectlArgs...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get chartmuseum service: %w", err)
	}

	serviceIP := strings.Trim(string(output), "'")
	if serviceIP == "" {
		return "", fmt.Errorf("chartmuseum service not found in namespace %s", namespace)
	}

	return fmt.Sprintf("http://%s:8080", serviceIP), nil
}

// publishChartToRepo publishes a chart to the repository
func publishChartToRepo(chartPath, repoURL string) error {
	// Read chart data
	chartData, err := os.ReadFile(chartPath)
	if err != nil {
		return fmt.Errorf("failed to read chart file: %w", err)
	}

	// First try to check if the chartmuseum API is available
	checkClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := checkClient.Get(repoURL + "/health")
	if err != nil {
		fmt.Printf("Warning: Chart repository health check failed: %v\n", err)
		fmt.Printf("Will still attempt to push chart, but it may fail.\n")
	} else {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Warning: Chart repository returned status %d for health check.\n", resp.StatusCode)
		} else {
			fmt.Printf("Chart repository is healthy.\n")
		}
	}

	// Create HTTP client
	client := &http.Client{Timeout: 30 * time.Second}

	// Create URL for the upload
	uploadURL := fmt.Sprintf("%s/api/charts", repoURL)

	fmt.Printf("Uploading chart to %s\n", uploadURL)

	// Create multipart form data for the upload
	// This is more compatible with chartmuseum
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file
	part, err := writer.CreateFormFile("chart", filepath.Base(chartPath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// Write chart data to form file
	if _, err := part.Write(chartData); err != nil {
		return fmt.Errorf("failed to write chart data: %w", err)
	}

	// Close multipart writer
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Perform the request
	resp, err = client.Do(req)
	if err != nil {
		// Try an alternative approach with direct chart upload
		return tryDirectChartUpload(chartPath, uploadURL)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("Warning: Upload failed with status %d: %s\n", resp.StatusCode, string(bodyBytes))
		return tryDirectChartUpload(chartPath, uploadURL)
	}

	fmt.Printf("Successfully published chart: %s\n", filepath.Base(chartPath))
	return nil
}

// tryDirectChartUpload attempts to upload the chart directly using application/gzip content-type
func tryDirectChartUpload(chartPath, uploadURL string) error {
	fmt.Printf("Trying alternative upload method for %s...\n", filepath.Base(chartPath))

	// Read chart data
	chartData, err := os.ReadFile(chartPath)
	if err != nil {
		return fmt.Errorf("failed to read chart file: %w", err)
	}

	// Create HTTP client
	client := &http.Client{Timeout: 30 * time.Second}

	// Create request
	req, err := http.NewRequest("POST", uploadURL, bytes.NewBuffer(chartData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/gzip")

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload chart (alternative method): %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	// Check response status
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("Manual chart upload steps:\n")
		fmt.Printf("1. Use 'curl' to upload the chart:\n")
		fmt.Printf("   curl -X POST -F 'chart=@%s' %s\n", chartPath, uploadURL)
		return fmt.Errorf("failed to upload chart, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Printf("Successfully published chart with alternative method: %s\n", filepath.Base(chartPath))
	return nil
}

func init() {
	// Initialize registry command
	registryCmd := &cobra.Command{
		Use:   "registry",
		Short: "Deploy a standalone Docker registry in a Kubernetes cluster",
		Long: `Deploy a standalone Docker registry in a Kubernetes cluster.
This command sets up a Docker registry that can be used for air-gapped 
Kubernetes deployments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flag values
			namespace, _ := cmd.Flags().GetString("namespace")
			image, _ := cmd.Flags().GetString("image")
			persistent, _ := cmd.Flags().GetBool("persistent")
			kubeconfigPath, _ := cmd.Flags().GetString("kubeconfig")
			localBuild, _ := cmd.Flags().GetBool("local-build")

			// If local-build is specified, build a local registry image
			if localBuild {
				fmt.Println("Building a local registry image for air-gapped deployment...")
				if err := buildLocalRegistryImage(); err != nil {
					return fmt.Errorf("failed to build local registry image: %w", err)
				}
				// Update the image path to use the local image
				image = "localhost:5000/registry:local"
			}

			return runRegistry(namespace, image, persistent, kubeconfigPath)
		},
	}

	// Add flags to registry command
	registryCmd.Flags().String("namespace", "capsailer-registry", "Kubernetes namespace for the registry")
	registryCmd.Flags().String("image", "registry:2", "Container image for the registry")
	registryCmd.Flags().Bool("persistent", true, "Use persistent storage for the registry")
	registryCmd.Flags().String("kubeconfig", "", "Path to kubeconfig file")
	registryCmd.Flags().Bool("local-build", false, "Build a local registry image for air-gapped deployment")

	// Initialize push command
	pushCmd := &cobra.Command{
		Use:   "push",
		Short: "Push container images to the registry",
		Long: `Push container images to the local registry.
This command tags and pushes Docker images to the registry deployed with 'capsailer registry'.
You can push a single image or all images from a bundle.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flag values
			image, _ := cmd.Flags().GetString("image")
			bundlePath, _ := cmd.Flags().GetString("bundle")
			namespace, _ := cmd.Flags().GetString("namespace")
			kubeconfigPath, _ := cmd.Flags().GetString("kubeconfig")
			externalRegistry, _ := cmd.Flags().GetString("external-registry")
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")

			if image == "" && bundlePath == "" && externalRegistry == "" {
				return fmt.Errorf("either --image, --bundle, or --external-registry must be specified")
			}

			return runPush(image, bundlePath, namespace, kubeconfigPath, externalRegistry, username, password)
		},
	}

	// Add flags to push command
	pushCmd.Flags().String("image", "", "Container image to push (e.g., nginx:latest)")
	pushCmd.Flags().String("bundle", "", "Path to a Capsailer bundle file or directory to push all images from")
	pushCmd.Flags().String("namespace", "capsailer-registry", "Kubernetes namespace where the registry is deployed")
	pushCmd.Flags().String("kubeconfig", "", "Path to kubeconfig file")
	pushCmd.Flags().String("external-registry", "", "External registry to push images to (e.g., artifactory.example.com)")
	pushCmd.Flags().String("username", "", "Username for authentication with external registry")
	pushCmd.Flags().String("password", "", "Password for authentication with external registry")
	// Either image or bundle must be specified, but not marking either as required individually

	// Add commands to root
	rootCmd.AddCommand(registryCmd)
	rootCmd.AddCommand(pushCmd)

	// ... existing init() code ...
}

// buildLocalRegistryImage builds a registry image locally using a Dockerfile
func buildLocalRegistryImage() error {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "registry-build")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing temp directory: %v\n", err)
		}
	}()

	// Create a simple Dockerfile that just pulls the registry image
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	dockerfileContent := `FROM registry:2
# No changes needed, just using this to create a local copy
LABEL maintainer="Capsailer"
`

	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Build the image
	fmt.Println("Building registry image...")
	buildCmd := exec.Command("docker", "build", "-t", "localhost:5000/registry:local", tmpDir)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build registry image: %w", err)
	}

	fmt.Println("Registry image built successfully: localhost:5000/registry:local")
	return nil
}
