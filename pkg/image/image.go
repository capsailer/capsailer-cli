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
	// This is a placeholder that will be implemented later
	// It would use the go-containerregistry library to load the image
	// into a registry running in the air-gapped environment
	
	return fmt.Errorf("not implemented yet")
}

// ReplaceImageReferences updates image references in Helm values to use the local registry
func ReplaceImageReferences(valuesFile string, registryHost string, imageMap map[string]string) error {
	// This is a placeholder that will be implemented later
	// It would parse the YAML values file and replace image references
	// with ones pointing to the local registry
	
	return fmt.Errorf("not implemented yet")
}

// sanitizeFilename converts an image name to a safe filename
func sanitizeFilename(imageName string) string {
	// Replace invalid characters with underscores
	name := strings.ReplaceAll(imageName, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "@", "_")
	
	return name
} 