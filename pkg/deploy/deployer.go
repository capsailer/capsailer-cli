package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jhennig/capsailer/pkg/helm"
	"gopkg.in/yaml.v3"
)

// DeployOptions defines options for deployment
type DeployOptions struct {
	ChartName   string
	ValuesFile  string
	Namespace   string
	ReleaseName string
	Registry    string
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

	return &Deployer{
		Options: options,
	}
}

// Deploy deploys a Helm chart
func (d *Deployer) Deploy() error {
	// Find the chart
	chartPath, err := findChart(d.Options.ChartName)
	if err != nil {
		return fmt.Errorf("failed to find chart: %w", err)
	}

	// Load the chart
	_, err = helm.LoadChart(chartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
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

// findChart looks for a chart in the local charts directory
func findChart(name string) (string, error) {
	// Check if chart exists in charts directory
	chartPath := filepath.Join("charts", name)
	if _, err := os.Stat(chartPath); err == nil {
		return chartPath, nil
	}

	// Try with .tgz extension
	chartPath = filepath.Join("charts", name + ".tgz")
	if _, err := os.Stat(chartPath); err == nil {
		return chartPath, nil
	}

	// Try finding by pattern
	matches, err := filepath.Glob(filepath.Join("charts", name + "-*.tgz"))
	if err == nil && len(matches) > 0 {
		// Use the first match
		return matches[0], nil
	}

	return "", fmt.Errorf("chart '%s' not found in charts directory", name)
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
			rewriteImageReferences(subMap, registry)
		}
	}

	return nil
} 