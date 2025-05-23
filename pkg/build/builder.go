package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/capsailer/capsailer-cli/pkg/helm"
	"github.com/capsailer/capsailer-cli/pkg/utils"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

// BuildOptions defines options for the build process
type BuildOptions struct {
	ManifestPath           string
	OutputPath             string
	Parallel               int
	RewriteImageReferences bool
	RegistryURL            string
}

// Builder handles the build process
type Builder struct {
	options BuildOptions
	tracker *utils.ProgressTracker
}

// NewBuilder creates a new Builder with the given options
func NewBuilder(options BuildOptions) *Builder {
	return &Builder{
		options: options,
		tracker: utils.NewProgressTracker(),
	}
}

// Build builds a bundle from the manifest
func (b *Builder) Build() error {
	// Create temporary directory for build
	tempDir, err := os.MkdirTemp("", "capsailer-build-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing temp directory: %v\n", err)
		}
	}()

	// Load and validate the manifest
	manifest, err := utils.LoadManifest(b.options.ManifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Create directory structure
	imagesDir := filepath.Join(tempDir, "images")
	chartsDir := filepath.Join(tempDir, "charts")

	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create images directory: %w", err)
	}

	if err := os.MkdirAll(chartsDir, 0755); err != nil {
		return fmt.Errorf("failed to create charts directory: %w", err)
	}

	// Download images
	fmt.Println("Downloading images...")
	if err := b.downloadImages(manifest.Images, imagesDir); err != nil {
		return fmt.Errorf("failed to download images: %w", err)
	}

	// Download charts
	fmt.Println("Downloading charts...")
	if err := b.downloadCharts(manifest.Charts, chartsDir); err != nil {
		return fmt.Errorf("failed to download charts: %w", err)
	}

	// Rewrite image references in charts if requested
	if b.options.RewriteImageReferences {
		if b.options.RegistryURL == "" {
			return fmt.Errorf("registry URL is required when rewriting image references")
		}

		fmt.Println("Rewriting image references in Helm charts...")
		if err := b.rewriteImageReferencesInCharts(manifest.Charts, chartsDir); err != nil {
			return fmt.Errorf("failed to rewrite image references: %w", err)
		}
	}

	// Copy values files
	if err := b.copyValuesFiles(manifest.Charts, chartsDir); err != nil {
		return fmt.Errorf("failed to copy values files: %w", err)
	}

	// Copy manifest to temp directory
	manifestData, err := os.ReadFile(b.options.ManifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "manifest.yaml"), manifestData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest to temp directory: %w", err)
	}

	// Create bundle
	fmt.Println("Creating bundle...")
	if err := utils.CreateTarGz(tempDir, b.options.OutputPath, b.tracker); err != nil {
		return fmt.Errorf("failed to create bundle: %w", err)
	}

	fmt.Printf("Bundle created successfully: %s\n", b.options.OutputPath)
	return nil
}

// downloadImages downloads container images in parallel
func (b *Builder) downloadImages(images []string, outputDir string) error {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, b.options.Parallel)
	errChan := make(chan error, len(images))

	for _, image := range images {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(img string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			if err := b.downloadImage(img, outputDir); err != nil {
				errChan <- fmt.Errorf("failed to download image %s: %w", img, err)
			}
		}(image)
	}

	// Wait for all downloads to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	if len(errChan) > 0 {
		err := <-errChan
		return err
	}

	// Add a small delay to ensure all progress bars are properly rendered
	time.Sleep(100 * time.Millisecond)

	return nil
}

// downloadImage downloads a single container image using go-containerregistry
func (b *Builder) downloadImage(image, outputDir string) error {
	// Parse the image reference
	ref, err := name.ParseReference(image)
	if err != nil {
		return fmt.Errorf("failed to parse image reference: %w", err)
	}

	// Pull the image
	img, err := remote.Image(ref, remote.WithContext(context.Background()))
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Get image size
	size, err := img.Size()
	if err != nil {
		return fmt.Errorf("failed to get image size: %w", err)
	}

	// Create a filename-safe version of the image name
	safeImageName := strings.ReplaceAll(image, "/", "_")
	safeImageName = strings.ReplaceAll(safeImageName, ":", "_")
	outputPath := filepath.Join(outputDir, safeImageName+".tar")

	// Add progress bar
	b.tracker.AddProgressBar(image, size)

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Save the image as a tarball
	if err := tarball.WriteToFile(outputPath, ref, img); err != nil {
		b.tracker.Finish(image) // Ensure we clean up the progress bar on error
		return fmt.Errorf("failed to save image: %w", err)
	}

	// Update progress to 100%
	b.tracker.Increment(image, size)

	// Mark progress as complete
	b.tracker.Finish(image)

	return nil
}

// downloadCharts downloads Helm charts
func (b *Builder) downloadCharts(charts []utils.Chart, outputDir string) error {
	for _, chart := range charts {
		// Create a chart repository
		repoURL := chart.Repo
		repoName := fmt.Sprintf("capsailer-%s", chart.Name)

		// Create temp directory for repo cache
		cacheDir, err := os.MkdirTemp("", "capsailer-helm-cache")
		if err != nil {
			return fmt.Errorf("failed to create temp directory for helm cache: %w", err)
		}
		defer func() {
			if err := os.RemoveAll(cacheDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error removing cache directory: %v\n", err)
			}
		}()

		// Initialize the chart repository and download the index file
		if err := initChartRepo(repoName, repoURL, cacheDir); err != nil {
			return fmt.Errorf("failed to initialize chart repository: %w", err)
		}

		// Find the chart version in the index
		indexPath := filepath.Join(cacheDir, fmt.Sprintf("%s-index.yaml", repoName))
		indexFile, err := repo.LoadIndexFile(indexPath)
		if err != nil {
			return fmt.Errorf("failed to load repository index: %w", err)
		}

		chartVersions, ok := indexFile.Entries[chart.Name]
		if !ok {
			return fmt.Errorf("chart %s not found in repository", chart.Name)
		}

		// Find the requested version
		var chartURL string
		for _, ver := range chartVersions {
			if ver.Version == chart.Version {
				if len(ver.URLs) == 0 {
					return fmt.Errorf("no download URL found for chart %s version %s", chart.Name, chart.Version)
				}
				chartURL = ver.URLs[0]
				break
			}
		}

		if chartURL == "" {
			return fmt.Errorf("chart version %s not found for %s", chart.Version, chart.Name)
		}

		// If URL is relative, prepend the repo URL
		if !strings.HasPrefix(chartURL, "http://") && !strings.HasPrefix(chartURL, "https://") {
			chartURL = strings.TrimSuffix(repoURL, "/") + "/" + strings.TrimPrefix(chartURL, "/")
		}

		// Add progress bar
		chartName := fmt.Sprintf("%s-%s", chart.Name, chart.Version)
		b.tracker.AddProgressBar(chartName, 100) // We'll update this in chunks

		// Set up HTTP getter
		getters := getter.Providers{
			getter.Provider{
				Schemes: []string{"http", "https"},
				New:     getter.NewHTTPGetter,
			},
		}

		// Get an HTTP client
		httpGetter, err := getters.ByScheme("https")
		if err != nil {
			return fmt.Errorf("failed to get HTTP getter: %w", err)
		}

		// Download the chart
		chartFileName := fmt.Sprintf("%s-%s.tgz", chart.Name, chart.Version)
		outputPath := filepath.Join(outputDir, chartFileName)

		// Download the chart data
		data, err := httpGetter.Get(chartURL)
		if err != nil {
			return fmt.Errorf("failed to download chart: %w", err)
		}

		// Write chart data to file
		if err := os.WriteFile(outputPath, data.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write chart file: %w", err)
		}

		// Update progress to 100%
		b.tracker.Increment(chartName, 100)

		// Mark progress as complete
		b.tracker.Finish(chartName)
	}

	return nil
}

// Helper function to initialize a chart repository
func initChartRepo(name, url, cacheDir string) error {
	// Create repo entry
	entry := &repo.Entry{
		Name: name,
		URL:  url,
	}

	// Create providers
	providers := getter.Providers{
		getter.Provider{
			Schemes: []string{"http", "https"},
			New:     getter.NewHTTPGetter,
		},
	}

	// Create chart repository
	chartRepo, err := repo.NewChartRepository(entry, providers)
	if err != nil {
		return err
	}

	// Set cache path
	chartRepo.CachePath = cacheDir

	// Download the index file
	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		return fmt.Errorf("failed to download repository index: %w", err)
	}

	return nil
}

// copyValuesFiles copies Helm chart values files
func (b *Builder) copyValuesFiles(charts []utils.Chart, outputDir string) error {
	for _, chart := range charts {
		if chart.ValuesFile == "" {
			continue
		}

		fmt.Printf("Copying values file for chart: %s\n", chart.Name)

		// Read values file
		valuesData, err := os.ReadFile(chart.ValuesFile)
		if err != nil {
			return fmt.Errorf("failed to read values file %s: %w", chart.ValuesFile, err)
		}

		// Write to output directory
		outputPath := filepath.Join(outputDir, filepath.Base(chart.ValuesFile))
		if err := os.WriteFile(outputPath, valuesData, 0644); err != nil {
			return fmt.Errorf("failed to write values file: %w", err)
		}

		fmt.Printf("Copied values file: %s\n", outputPath)
	}

	return nil
}

// rewriteImageReferencesInCharts rewrites image references in all charts
func (b *Builder) rewriteImageReferencesInCharts(charts []utils.Chart, chartsDir string) error {
	for _, chart := range charts {
		fmt.Printf("Rewriting image references in chart: %s\n", chart.Name)

		// Find the chart file
		chartPath := filepath.Join(chartsDir, fmt.Sprintf("%s-%s.tgz", chart.Name, chart.Version))
		if _, err := os.Stat(chartPath); os.IsNotExist(err) {
			// Try alternative naming patterns
			pattern := filepath.Join(chartsDir, fmt.Sprintf("%s-*.tgz", chart.Name))
			matches, err := filepath.Glob(pattern)
			if err != nil || len(matches) == 0 {
				return fmt.Errorf("chart file not found for %s-%s", chart.Name, chart.Version)
			}
			chartPath = matches[0]
		}

		// Rewrite image references
		if err := helm.RewriteImageReferences(chartPath, b.options.RegistryURL); err != nil {
			return fmt.Errorf("failed to rewrite image references in chart %s: %w", chart.Name, err)
		}
	}

	return nil
}
