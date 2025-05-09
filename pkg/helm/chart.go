package helm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

// ChartInfo stores information about a Helm chart
type ChartInfo struct {
	Name    string
	Version string
	Path    string
}

// DownloadChart downloads a Helm chart from a repository
func DownloadChart(name, repoURL, version, outputDir string) (*ChartInfo, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Initialize settings with default values
	settings := cli.New()

	// Create a temporary directory for the repository cache
	tempDir, err := ioutil.TempDir("", "capsailer-helm-cache")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Update repository settings
	repoCache := filepath.Join(tempDir, "repository")
	if err := os.MkdirAll(repoCache, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repository cache directory: %w", err)
	}

	// Initialize the chart repository
	chartRepo := &repo.Entry{
		Name: "temp-repo",
		URL:  repoURL,
	}

	// Create chart repository
	r, err := repo.NewChartRepository(chartRepo, getter.All(settings))
	if err != nil {
		return nil, fmt.Errorf("failed to create chart repository: %w", err)
	}

	// Set repository cache
	r.CachePath = repoCache

	// Download the repository index
	if _, err := r.DownloadIndexFile(); err != nil {
		return nil, fmt.Errorf("failed to download repository index: %w", err)
	}

	// Initialize chart downloader
	dl := downloader.ChartDownloader{
		Out:              os.Stdout,
		Keyring:          "",
		Getters:          getter.All(settings),
		RepositoryConfig: "",
		RepositoryCache:  repoCache,
	}

	// Download the chart
	chartPath, _, err := dl.DownloadTo(fmt.Sprintf("%s/%s", chartRepo.Name, name), version, outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to download chart: %w", err)
	}

	return &ChartInfo{
		Name:    name,
		Version: version,
		Path:    chartPath,
	}, nil
}

// LoadChart loads a Helm chart from a file
func LoadChart(chartPath string) (*chart.Chart, error) {
	return loader.Load(chartPath)
}

// InstallChart installs a Helm chart in Kubernetes
func InstallChart(chartPath, releaseName, namespace string, values map[string]interface{}) error {
	// This is a placeholder that will be implemented later
	// It would use the Helm SDK to install a chart in the Kubernetes cluster
	
	return fmt.Errorf("not implemented yet")
} 