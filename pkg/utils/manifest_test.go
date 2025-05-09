package utils

import (
	"os"
	"testing"
)

func TestValidateManifest(t *testing.T) {
	// Test with an empty manifest
	emptyManifest := &Manifest{}
	err := validateManifest(emptyManifest)
	if err == nil {
		t.Error("Expected error for empty manifest, but got nil")
	}

	// Test with a valid manifest
	validManifest := &Manifest{
		Images: []string{"nginx:latest"},
		Charts: []Chart{
			{
				Name:    "test-chart",
				Repo:    "https://charts.example.com",
				Version: "1.0.0",
			},
		},
	}
	err = validateManifest(validManifest)
	if err != nil {
		t.Errorf("Expected no error for valid manifest, but got: %v", err)
	}
}

func TestLoadManifest(t *testing.T) {
	// Create a temporary manifest file for testing
	manifestContent := `
images:
  - nginx:latest
  - alpine:3.19
charts:
  - name: test-chart
    repo: https://charts.example.com
    version: 1.0.0
`
	tempFile, err := os.CreateTemp("", "test-manifest-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(manifestContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test loading the manifest
	manifest, err := LoadManifest(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	// Verify the loaded manifest
	if len(manifest.Images) != 2 {
		t.Errorf("Expected 2 images, got %d", len(manifest.Images))
	}
	if len(manifest.Charts) != 1 {
		t.Errorf("Expected 1 chart, got %d", len(manifest.Charts))
	}
	if manifest.Images[0] != "nginx:latest" {
		t.Errorf("Expected first image to be nginx:latest, got %s", manifest.Images[0])
	}
	if manifest.Charts[0].Name != "test-chart" {
		t.Errorf("Expected chart name to be test-chart, got %s", manifest.Charts[0].Name)
	}
} 