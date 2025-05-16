package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CreateTarGz creates a tar.gz archive from a source directory
func CreateTarGz(sourceDir, outputPath string, tracker *ProgressTracker) error {
	// Calculate total size first
	var totalSize int64
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to calculate total size: %w", err)
	}

	// Add progress bar
	tracker.AddProgressBar("Creating bundle", totalSize)

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file: %v\n", err)
		}
	}()

	// Create gzip writer
	gw := gzip.NewWriter(file)
	defer func() {
		if err := gw.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing gzip writer: %v\n", err)
		}
	}()

	// Create tar writer
	tw := tar.NewWriter(gw)
	defer func() {
		if err := tw.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing tar writer: %v\n", err)
		}
	}()

	// Walk through the source directory
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get header info
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return fmt.Errorf("failed to create tar header: %w", err)
		}

		// Update header name to be relative to source directory
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		header.Name = relPath

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// If it's a regular file, write content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file '%s': %w", path, err)
			}
			defer func() {
				if err := file.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "Error closing file: %v\n", err)
				}
			}()

			// Create a progress writer
			progressWriter := NewProgressWriter(tw, tracker, "Creating bundle")

			if _, err := io.Copy(progressWriter, file); err != nil {
				return fmt.Errorf("failed to write file to tar: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	// Mark progress as complete
	tracker.Finish("Creating bundle")

	return nil
}
