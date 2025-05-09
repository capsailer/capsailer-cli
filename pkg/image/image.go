package image

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"gopkg.in/yaml.v3"
)

// ImageInfo stores information about a container image
type ImageInfo struct {
	Name   string
	Digest string
	Size   int64
}

// PullImage pulls a container image and saves it to a tar file
func PullImage(imageName string, outputDir string) (*ImageInfo, error) {
	ref, err := name.ParseReference(imageName)
	if err != nil {
		return nil, fmt.Errorf("invalid image name '%s': %w", imageName, err)
	}

	// Pull the image from remote
	img, err := remote.Image(ref, remote.WithContext(context.Background()))
	if err != nil {
		return nil, fmt.Errorf("failed to pull image '%s': %w", imageName, err)
	}

	// Get the image digest
	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("failed to get digest for image '%s': %w", imageName, err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate safe filename from image name
	filename := sanitizeFilename(imageName) + ".tar"
	outputPath := filepath.Join(outputDir, filename)

	// Save the image as a tarball
	tag, err := name.NewTag(imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tag for image '%s': %w", imageName, err)
	}

	if err := tarball.WriteToFile(outputPath, tag, img); err != nil {
		return nil, fmt.Errorf("failed to save image '%s' to file: %w", imageName, err)
	}

	// Get size of the written file
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for '%s': %w", outputPath, err)
	}

	return &ImageInfo{
		Name:   imageName,
		Digest: digest.String(),
		Size:   fileInfo.Size(),
	}, nil
}

// LoadImage loads an image from a tar file into a registry
func LoadImage(imagePath string, registryHost string) error {
	// Load the image from the tar file
	img, err := tarball.ImageFromPath(imagePath, nil)
	if err != nil {
		return fmt.Errorf("failed to load image from %s: %w", imagePath, err)
	}

	// Extract the original tag from the tar filename
	basename := filepath.Base(imagePath)
	basename = strings.TrimSuffix(basename, ".tar")
	
	// Create a tag for the target registry
	tagName := fmt.Sprintf("%s/%s", registryHost, basename)
	newRef, err := name.NewTag(tagName)
	if err != nil {
		return fmt.Errorf("failed to create new tag: %w", err)
	}

	// Push the image to the registry
	if err := remote.Write(newRef, img); err != nil {
		return fmt.Errorf("failed to push image to registry: %w", err)
	}

	return nil
}

// ReplaceImageReferences updates image references in Helm values to use the local registry
func ReplaceImageReferences(valuesFile string, registryHost string, imageMap map[string]string) error {
	// Read the YAML file
	data, err := os.ReadFile(valuesFile)
	if err != nil {
		return fmt.Errorf("failed to read values file: %w", err)
	}

	// Parse the YAML content
	var values map[string]interface{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Process image references recursively
	processValues(values, registryHost, imageMap)

	// Write back the modified YAML
	updatedData, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal updated YAML: %w", err)
	}

	if err := os.WriteFile(valuesFile, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated values file: %w", err)
	}

	return nil
}

// processValues recursively processes a values map to replace image references
func processValues(values map[string]interface{}, registryHost string, imageMap map[string]string) {
	// Common patterns for image references in Helm charts
	imageKeys := []string{"repository", "image", "image.repository"}

	for key, value := range values {
		// Check if this is an image reference
		for _, imgKey := range imageKeys {
			if key == imgKey && value != nil {
				if strValue, ok := value.(string); ok {
					// Replace with the reference to the local registry
					values[key] = replaceImageRef(strValue, registryHost, imageMap)
				}
			}
		}

		// Handle special case for image sections that have repository field
		if key == "image" {
			if imgMap, ok := value.(map[string]interface{}); ok {
				if repo, ok := imgMap["repository"].(string); ok {
					imgMap["repository"] = replaceImageRef(repo, registryHost, imageMap)
				}
			}
		}

		// Recursively process nested maps
		if subMap, ok := value.(map[string]interface{}); ok {
			processValues(subMap, registryHost, imageMap)
		} else if subArray, ok := value.([]interface{}); ok {
			// Process arrays that might contain maps
			for i, item := range subArray {
				if subMap, ok := item.(map[string]interface{}); ok {
					processValues(subMap, registryHost, imageMap)
					subArray[i] = subMap
				}
			}
		}
	}
}

// replaceImageRef replaces an image reference with one pointing to the local registry
func replaceImageRef(imageRef, registryHost string, imageMap map[string]string) string {
	// If we have a mapping for this exact image, use it
	if newRef, ok := imageMap[imageRef]; ok {
		return newRef
	}
	
	// Otherwise, try to construct a new reference based on the original
	parts := strings.Split(imageRef, "/")
	baseName := parts[len(parts)-1]
	return fmt.Sprintf("%s/%s", registryHost, baseName)
}

// sanitizeFilename converts an image name to a safe filename
func sanitizeFilename(imageName string) string {
	// Replace invalid characters with underscores
	name := strings.ReplaceAll(imageName, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "@", "_")
	
	return name
} 