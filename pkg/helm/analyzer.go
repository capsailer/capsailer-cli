package helm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/capsailer/capsailer-cli/pkg/utils"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// ImageReference represents a container image reference found in a Helm chart
type ImageReference struct {
	Chart       string // Chart name
	Path        string // Path to the field in values.yaml (e.g., "image.repository")
	Repository  string // Image repository
	Tag         string // Image tag
	FullPath    string // Full YAML path to the field
	ValueSource string // Source of the value (e.g., "values.yaml", "templates")
}

// AnalyzeChartForImages analyzes a Helm chart for container image references
func AnalyzeChartForImages(chartPath string) ([]ImageReference, error) {
	// Load the chart
	chartObj, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	var references []ImageReference

	// First, analyze values.yaml for image references
	valuesRefs, err := analyzeValuesForImages(chartObj)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze values.yaml: %w", err)
	}
	references = append(references, valuesRefs...)

	// Second, analyze templates for hardcoded image references
	templateRefs, err := analyzeTemplatesForImages(chartObj)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze templates: %w", err)
	}
	references = append(references, templateRefs...)

	return references, nil
}

// analyzeValuesForImages extracts image references from values.yaml
func analyzeValuesForImages(chartObj *chart.Chart) ([]ImageReference, error) {
	var references []ImageReference

	// Convert values to a map
	valuesMap := chartObj.Values

	// Find image references in values
	findImageReferences(valuesMap, "", chartObj.Name(), &references, "values.yaml")

	return references, nil
}

// findImageReferences recursively searches for image references in a nested map
func findImageReferences(data map[string]interface{}, path string, chartName string, references *[]ImageReference, source string) {
	// Common image field patterns
	repoFields := []string{"repository", "image", "registry"}
	tagFields := []string{"tag", "imageTag"}

	// Check if this map has both repository and tag fields
	var repository, tag string
	var hasRepo, hasTag bool

	for _, field := range repoFields {
		if repo, ok := data[field]; ok {
			if repoStr, isStr := repo.(string); isStr && repoStr != "" {
				repository = repoStr
				hasRepo = true
				break
			}
		}
	}

	for _, field := range tagFields {
		if tagVal, ok := data[field]; ok {
			if tagStr, isStr := tagVal.(string); isStr && tagStr != "" {
				tag = tagStr
				hasTag = true
				break
			}
		}
	}

	// If we found both repository and tag at this level, add a reference
	if hasRepo && hasTag {
		*references = append(*references, ImageReference{
			Chart:       chartName,
			Path:        path,
			Repository:  repository,
			Tag:         tag,
			FullPath:    path,
			ValueSource: source,
		})
	} else if hasRepo {
		// If we only found a repository, add it with an empty tag
		*references = append(*references, ImageReference{
			Chart:       chartName,
			Path:        path,
			Repository:  repository,
			Tag:         "",
			FullPath:    path,
			ValueSource: source,
		})
	}

	// Recursively check nested maps
	for key, value := range data {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			newPath := key
			if path != "" {
				newPath = path + "." + key
			}
			findImageReferences(nestedMap, newPath, chartName, references, source)
		}
	}
}

// analyzeTemplatesForImages extracts hardcoded image references from templates
func analyzeTemplatesForImages(chartObj *chart.Chart) ([]ImageReference, error) {
	var references []ImageReference

	// Regular expression to find image: references in YAML/templates
	// This is a simplified pattern and might need to be enhanced for complex cases
	imageRegex := regexp.MustCompile(`(?:image|Image):\s*["']?([^"'\s}]+):([^"'\s}]+)["']?`)

	// Loop through all templates
	for _, template := range chartObj.Templates {
		matches := imageRegex.FindAllStringSubmatch(string(template.Data), -1)
		for _, match := range matches {
			if len(match) >= 3 {
				references = append(references, ImageReference{
					Chart:       chartObj.Name(),
					Path:        template.Name,
					Repository:  match[1],
					Tag:         match[2],
					FullPath:    template.Name,
					ValueSource: "template",
				})
			}
		}
	}

	return references, nil
}

// AnalyzeChartsInManifest analyzes all charts in a manifest for image references
func AnalyzeChartsInManifest(manifest *utils.Manifest, chartsDir string) (map[string][]ImageReference, error) {
	result := make(map[string][]ImageReference)

	for _, chart := range manifest.Charts {
		// Find the chart file
		chartPath := filepath.Join(chartsDir, fmt.Sprintf("%s-%s.tgz", chart.Name, chart.Version))
		if _, err := os.Stat(chartPath); os.IsNotExist(err) {
			// Try alternative naming patterns
			pattern := filepath.Join(chartsDir, fmt.Sprintf("%s-*.tgz", chart.Name))
			matches, err := filepath.Glob(pattern)
			if err != nil || len(matches) == 0 {
				return nil, fmt.Errorf("chart file not found for %s-%s", chart.Name, chart.Version)
			}
			chartPath = matches[0]
		}

		// Analyze the chart
		references, err := AnalyzeChartForImages(chartPath)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze chart %s: %w", chart.Name, err)
		}

		result[chart.Name] = references
	}

	return result, nil
}

// FindImagesNotInManifest compares images found in charts with those in the manifest
func FindImagesNotInManifest(chartImages map[string][]ImageReference, manifestImages []string) []string {
	// Create a map of manifest images for quick lookup
	manifestImageMap := make(map[string]bool)
	for _, img := range manifestImages {
		manifestImageMap[img] = true
		// Also add the image without tag for partial matching
		parts := strings.Split(img, ":")
		if len(parts) > 1 {
			manifestImageMap[parts[0]] = true
		}
	}

	// Find images in charts that are not in the manifest
	var missingImages []string
	seenImages := make(map[string]bool)

	for _, references := range chartImages {
		for _, ref := range references {
			fullImage := ref.Repository
			if ref.Tag != "" {
				fullImage = fullImage + ":" + ref.Tag
			}

			// Check if this image or its repository is in the manifest
			_, inManifest := manifestImageMap[fullImage]
			_, repoInManifest := manifestImageMap[ref.Repository]

			if !inManifest && !repoInManifest && !seenImages[fullImage] {
				missingImages = append(missingImages, fullImage)
				seenImages[fullImage] = true
			}
		}
	}

	return missingImages
}

// RewriteImageReferences rewrites image references in a chart to use a private registry
func RewriteImageReferences(chartPath, registryURL string) error {
	// Create a temporary directory to extract the chart
	tempDir, err := os.MkdirTemp("", "chart-extract-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract the chart to the temp directory
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Extract the chart using tar command
	cmd := exec.Command("tar", "-xzf", chartPath, "-C", tempDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract chart: %w", err)
	}

	// Find the chart directory (should be only one directory)
	dirs, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	if len(dirs) == 0 {
		return fmt.Errorf("no chart found in the archive")
	}

	chartDir := filepath.Join(tempDir, dirs[0].Name())

	// Read and modify the values.yaml file
	valuesPath := filepath.Join(chartDir, "values.yaml")
	valuesData, err := os.ReadFile(valuesPath)
	if err != nil {
		return fmt.Errorf("failed to read values.yaml: %w", err)
	}

	// Parse values.yaml
	var values map[string]interface{}
	if err := yaml.Unmarshal(valuesData, &values); err != nil {
		return fmt.Errorf("failed to parse values.yaml: %w", err)
	}

	// Rewrite image references
	rewriteImageReferencesInValues(values, registryURL)

	// Convert back to YAML
	newValuesData, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal values.yaml: %w", err)
	}

	// Write the modified values.yaml
	if err := os.WriteFile(valuesPath, newValuesData, 0644); err != nil {
		return fmt.Errorf("failed to write values.yaml: %w", err)
	}

	// Create a new chart archive
	outputTempFile, err := os.CreateTemp("", "chart-*.tgz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	outputTempFile.Close()
	outputPath := outputTempFile.Name()

	// Create the archive using tar command
	cmd = exec.Command("tar", "-czf", outputPath, "-C", tempDir, dirs[0].Name())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create chart archive: %w", err)
	}

	// Replace the original chart
	if err := os.Rename(outputPath, chartPath); err != nil {
		return fmt.Errorf("failed to replace original chart: %w", err)
	}

	return nil
}

// rewriteImageReferencesInValues rewrites image references in values map
func rewriteImageReferencesInValues(data map[string]interface{}, registryURL string) {
	// Check if this is an image configuration section with registry and repository fields
	if _, hasRegistry := data["registry"].(string); hasRegistry {
		if repository, hasRepo := data["repository"].(string); hasRepo && repository != "" {
			// We found both registry and repository fields
			// Update the registry field to our private registry
			data["registry"] = registryURL
			
			// If the repository includes a path (e.g., bitnami/nginx), keep only the image name
			if strings.Contains(repository, "/") {
				parts := strings.Split(repository, "/")
				data["repository"] = parts[len(parts)-1]
			}
			
			return
		}
	}
	
	// Check for standalone repository fields
	repoFields := []string{"repository", "image"}
	for _, field := range repoFields {
		if repo, ok := data[field]; ok {
			if repoStr, isStr := repo.(string); isStr && repoStr != "" {
				// Don't rewrite if it's already using our registry
				if !strings.HasPrefix(repoStr, registryURL) {
					// Extract the image name without registry
					parts := strings.Split(repoStr, "/")
					imageName := parts[len(parts)-1]
					
					// Rewrite to use our registry
					data[field] = fmt.Sprintf("%s/%s", registryURL, imageName)
				}
			}
		}
	}

	// Recursively check nested maps
	for _, value := range data {
		if nestedMap, ok := value.(map[string]interface{}); ok {
			rewriteImageReferencesInValues(nestedMap, registryURL)
		}
	}
} 