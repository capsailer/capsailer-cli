package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndExtractTarGz(t *testing.T) {
	// Create temporary directories for test
	sourceDir, err := os.MkdirTemp("", "capsailer-test-source-")
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	defer os.RemoveAll(sourceDir)
	
	targetDir, err := os.MkdirTemp("", "capsailer-test-target-")
	if err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}
	defer os.RemoveAll(targetDir)
	
	// Create a test file in the source directory
	testFile := filepath.Join(sourceDir, "test.txt")
	testContent := "Hello, Capsailer!"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create a subdirectory with a file
	subDir := filepath.Join(sourceDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	subFile := filepath.Join(subDir, "subfile.txt")
	if err := os.WriteFile(subFile, []byte("Nested file"), 0644); err != nil {
		t.Fatalf("Failed to create subdir file: %v", err)
	}
	
	// Create the archive
	archivePath := filepath.Join(targetDir, "test-archive.tar.gz")
	if err := CreateTarGz(sourceDir, archivePath); err != nil {
		t.Fatalf("Failed to create tar.gz: %v", err)
	}
	
	// Check if the archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatalf("Archive file was not created")
	}
	
	// Extract the archive to a new location
	extractDir := filepath.Join(targetDir, "extracted")
	if err := os.Mkdir(extractDir, 0755); err != nil {
		t.Fatalf("Failed to create extraction directory: %v", err)
	}
	
	// Create an unpacker and extract
	unpacker := NewUnpacker(UnpackOptions{
		BundlePath: archivePath,
		OutputDir:  extractDir,
	})
	
	if err := unpacker.Unpack(); err != nil {
		t.Fatalf("Failed to extract archive: %v", err)
	}
	
	// Verify the extracted file contains the expected content
	extractedFile := filepath.Join(extractDir, "test.txt")
	content, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}
	
	if string(content) != testContent {
		t.Fatalf("Extracted content doesn't match. Got %q, expected %q", string(content), testContent)
	}
	
	// Verify the subdirectory and its file were extracted
	extractedSubFile := filepath.Join(extractDir, "subdir", "subfile.txt")
	if _, err := os.Stat(extractedSubFile); os.IsNotExist(err) {
		t.Fatalf("Subdirectory file was not extracted")
	}
} 