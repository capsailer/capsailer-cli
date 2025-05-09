package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadManifest loads and validates a manifest from a file
func LoadManifest(filePath string) (*Manifest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	manifest := &Manifest{}
	if err := yaml.Unmarshal(data, manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}

	if err := validateManifest(manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}

// validateManifest ensures the manifest contains valid entries
func validateManifest(manifest *Manifest) error {
	if len(manifest.Images) == 0 && len(manifest.Charts) == 0 {
		return errors.New("manifest must contain at least one image or chart")
	}

	// Validate images
	for i, image := range manifest.Images {
		if strings.TrimSpace(image) == "" {
			return fmt.Errorf("image at index %d is empty", i)
		}
		// TODO: More advanced image validation could be added here
	}

	// Validate charts
	for i, chart := range manifest.Charts {
		if strings.TrimSpace(chart.Name) == "" {
			return fmt.Errorf("chart at index %d has no name", i)
		}
		if strings.TrimSpace(chart.Repo) == "" {
			return fmt.Errorf("chart at index %d has no repository", i)
		}
		if strings.TrimSpace(chart.Version) == "" {
			return fmt.Errorf("chart at index %d has no version", i)
		}
		if chart.ValuesFile != "" {
			if _, err := os.Stat(chart.ValuesFile); os.IsNotExist(err) {
				return fmt.Errorf("values file '%s' for chart '%s' does not exist", chart.ValuesFile, chart.Name)
			}
		}
	}

	return nil
}

// SaveManifest writes a manifest to a file
func SaveManifest(manifest *Manifest, filePath string) error {
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest to YAML: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
} 