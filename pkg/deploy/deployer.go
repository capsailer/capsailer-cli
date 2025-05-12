package deploy

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlnhnng/capsailer/pkg/helm"
	yaml "gopkg.in/yaml.v3"
)

// DeployOptions defines options for deployment
type DeployOptions struct {
	ChartName      string
	ValuesFile     string
	Namespace      string
	ReleaseName    string
	Registry       string
	KubeconfigPath string
	RegistryNamespace string // Namespace where registry and chartmuseum are deployed
}

// Deployer handles deploying Helm charts
type Deployer struct {
	Options DeployOptions
}

// NewDeployer creates a new Deployer instance
func NewDeployer(options DeployOptions) *Deployer {
	// Set defaults if needed
	if options.Namespace == "" {
		options.Namespace = "default"
	}
	
	if options.ReleaseName == "" {
		options.ReleaseName = options.ChartName
	}
	
	if options.Registry == "" {
		options.Registry = "localhost:5000"
	}

	if options.RegistryNamespace == "" {
		options.RegistryNamespace = "capsailer-registry"
	}

	return &Deployer{
		Options: options,
	}
}

// Deploy deploys a Helm chart
func (d *Deployer) Deploy() error {
	// Find the chart locally or in ChartMuseum
	chartPath, isLocal, err := d.findChart(d.Options.ChartName)
	if err != nil {
		return fmt.Errorf("failed to find chart: %w", err)
	}

	var chartObj *helm.ChartInfo

	// If the chart is not local, we need to download it from ChartMuseum
	if !isLocal {
		fmt.Printf("Chart '%s' found in ChartMuseum, downloading...\n", d.Options.ChartName)
		
		// Set up port forwarding to ChartMuseum
		forwardCmd := exec.Command("kubectl", "port-forward", "-n", d.Options.RegistryNamespace, "svc/chartmuseum", "8080:8080")
		if d.Options.KubeconfigPath != "" {
			forwardCmd.Args = append(forwardCmd.Args, "--kubeconfig", d.Options.KubeconfigPath)
		}
		
		// Start port-forwarding in background
		if err := forwardCmd.Start(); err != nil {
			return fmt.Errorf("failed to start port-forward to ChartMuseum: %w", err)
		}
		
		// Ensure we stop the port-forwarding when done
		defer func() {
			if forwardCmd.Process != nil {
				if err := forwardCmd.Process.Kill(); err != nil {
					fmt.Fprintf(os.Stderr, "Error stopping port forwarding: %v\n", err)
				}
			}
		}()
		
		// Give port-forwarding time to establish
		fmt.Println("Setting up port-forwarding to ChartMuseum for download...")
		time.Sleep(2 * time.Second)
		
		// Use localhost URL for chart download
		repoURL := "http://localhost:8080"

		// Download the chart to a temporary directory
		tempDir, err := os.MkdirTemp("", "capsailer-chart-")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error removing temp directory: %v\n", err)
			}
		}()

		// Download the chart
		chartObj, err = helm.DownloadChart(d.Options.ChartName, repoURL, "", tempDir)
		if err != nil {
			return fmt.Errorf("failed to download chart from ChartMuseum: %w", err)
		}
		chartPath = chartObj.Path
	} else {
		// Load the chart
		_, err = helm.LoadChart(chartPath)
		if err != nil {
			return fmt.Errorf("failed to load chart: %w", err)
		}
	}

	// Load values file
	values, err := loadValues(d.Options.ValuesFile)
	if err != nil {
		return fmt.Errorf("failed to load values: %w", err)
	}

	// Rewrite image references in values to use the local registry
	if err := rewriteImageReferences(values, d.Options.Registry); err != nil {
		return fmt.Errorf("failed to rewrite image references: %w", err)
	}

	// Install the chart
	fmt.Printf("Installing chart %s as release %s in namespace %s\n", 
		d.Options.ChartName, d.Options.ReleaseName, d.Options.Namespace)
	
	if err := helm.InstallChart(chartPath, d.Options.ReleaseName, d.Options.Namespace, values); err != nil {
		return fmt.Errorf("failed to install chart: %w", err)
	}

	fmt.Printf("Successfully deployed %s\n", d.Options.ChartName)
	return nil
}

// findChart looks for a chart in the local charts directory or in ChartMuseum
func (d *Deployer) findChart(name string) (string, bool, error) {
	// First, check if chart exists in local charts directory
	chartPath := filepath.Join("charts", name)
	if _, err := os.Stat(chartPath); err == nil {
		return chartPath, true, nil
	}

	// Try with .tgz extension
	chartPath = filepath.Join("charts", name + ".tgz")
	if _, err := os.Stat(chartPath); err == nil {
		return chartPath, true, nil
	}

	// Try finding by pattern
	matches, err := filepath.Glob(filepath.Join("charts", name + "-*.tgz"))
	if err == nil && len(matches) > 0 {
		// Use the first match
		return matches[0], true, nil
	}

	// If not found locally, check if ChartMuseum is available
	fmt.Println("Chart not found locally, checking if ChartMuseum is available...")
	
	// Check if ChartMuseum is running in the cluster
	if chartExists, err := d.chartExistsInChartMuseum(name); err == nil && chartExists {
		// Return a placeholder path and isLocal=false to indicate it's in ChartMuseum
		return name, false, nil
	} else if err != nil {
		fmt.Printf("Warning: Failed to check ChartMuseum: %v\n", err)
	}

	return "", false, fmt.Errorf("chart '%s' not found locally or in ChartMuseum", name)
}

// chartExistsInChartMuseum checks if a chart exists in ChartMuseum
func (d *Deployer) chartExistsInChartMuseum(chartName string) (bool, error) {
	// Get ChartMuseum URL
	repoURL, err := d.getChartMuseumURL()
	if err != nil {
		return false, fmt.Errorf("failed to get ChartMuseum URL: %w", err)
	}

	// Check if ChartMuseum is available by querying its API
	// We'll use port-forwarding to access it, so we don't need this URL directly
	_ = fmt.Sprintf("%s/api/charts/%s", repoURL, chartName)
	
	// Use kubectl port-forward to access ChartMuseum
	forwardCmd := exec.Command("kubectl", "port-forward", "-n", d.Options.RegistryNamespace, "svc/chartmuseum", "8080:8080")
	if d.Options.KubeconfigPath != "" {
		forwardCmd.Args = append(forwardCmd.Args, "--kubeconfig", d.Options.KubeconfigPath)
	}
	
	// Start port-forwarding in background
	if err := forwardCmd.Start(); err != nil {
		return false, fmt.Errorf("failed to start port-forward to ChartMuseum: %w", err)
	}
	
	// Ensure we stop the port-forwarding when done
	defer func() {
		if forwardCmd.Process != nil {
			if err := forwardCmd.Process.Kill(); err != nil {
				fmt.Fprintf(os.Stderr, "Error stopping port forwarding: %v\n", err)
			}
		}
	}()
	
	// Give port-forwarding time to establish
	fmt.Println("Setting up port-forwarding to ChartMuseum...")
	
	// Use the local port-forwarded URL
	localURL := "http://localhost:8080/api/charts/" + chartName
	
	// Wait a moment for port-forwarding to establish
	time.Sleep(2 * time.Second)
	
	// Check if chart exists
	resp, err := http.Get(localURL)
	if err != nil {
		return false, fmt.Errorf("failed to connect to ChartMuseum: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()
	
	// If status code is 200, chart exists
	return resp.StatusCode == http.StatusOK, nil
}

// getChartMuseumURL gets the URL for ChartMuseum
func (d *Deployer) getChartMuseumURL() (string, error) {
	// Get ChartMuseum service IP
	kubectlArgs := []string{"get", "service", "-n", d.Options.RegistryNamespace, "chartmuseum", "-o", "jsonpath={.spec.clusterIP}"}
	if d.Options.KubeconfigPath != "" {
		kubectlArgs = append(kubectlArgs, "--kubeconfig", d.Options.KubeconfigPath)
	}
	
	cmd := exec.Command("kubectl", kubectlArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get ChartMuseum service: %w", err)
	}
	
	serviceIP := strings.TrimSpace(string(output))
	if serviceIP == "" {
		return "", fmt.Errorf("ChartMuseum service not found in namespace %s", d.Options.RegistryNamespace)
	}
	
	return fmt.Sprintf("http://%s:8080", serviceIP), nil
}

// loadValues loads values from a YAML file
func loadValues(filename string) (map[string]interface{}, error) {
	if filename == "" {
		return map[string]interface{}{}, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read values file: %w", err)
	}

	var values map[string]interface{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("failed to parse values YAML: %w", err)
	}

	return values, nil
}

// rewriteImageReferences updates image references in values to use the local registry
func rewriteImageReferences(values map[string]interface{}, registry string) error {
	// This is a simplified version - a real implementation would need to be more sophisticated
	// to handle various chart structures

	// Look for common patterns of image references
	if image, ok := values["image"]; ok {
		if imageMap, ok := image.(map[string]interface{}); ok {
			if repository, ok := imageMap["repository"].(string); ok {
				// Rewrite the repository
				imageMap["repository"] = fmt.Sprintf("%s/%s", registry, filepath.Base(repository))
			}
		}
	}

	// Handle more complex structures recursively
	for _, value := range values {
		if subMap, ok := value.(map[string]interface{}); ok {
			if err := rewriteImageReferences(subMap, registry); err != nil {
				return err
			}
		}
	}

	return nil
} 