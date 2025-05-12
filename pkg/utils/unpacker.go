package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// UnpackOptions defines options for unpacking
type UnpackOptions struct {
	BundlePath string
	OutputDir  string
}

// Unpacker handles extracting capsailer bundles
type Unpacker struct {
	Options UnpackOptions
}

// NewUnpacker creates a new Unpacker instance
func NewUnpacker(options UnpackOptions) *Unpacker {
	// Set defaults if needed
	if options.OutputDir == "" {
		options.OutputDir = "."
	}

	return &Unpacker{
		Options: options,
	}
}

// Unpack extracts a bundle to a directory
func (u *Unpacker) Unpack() error {
	// Validate bundle path
	if u.Options.BundlePath == "" {
		return fmt.Errorf("bundle path is required")
	}

	// Check if bundle exists
	if _, err := os.Stat(u.Options.BundlePath); os.IsNotExist(err) {
		return fmt.Errorf("bundle file '%s' does not exist", u.Options.BundlePath)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(u.Options.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Open the tar.gz file
	file, err := os.Open(u.Options.BundlePath)
	if err != nil {
		return fmt.Errorf("failed to open bundle: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file: %v\n", err)
		}
	}()

	// Create a gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if err := gzr.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing gzip reader: %v\n", err)
		}
	}()

	// Create a tar reader
	tr := tar.NewReader(gzr)

	// Extract each file
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Get the target path
		target := filepath.Join(u.Options.OutputDir, header.Name)

		// Handle based on file type
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w", target, err)
			}

		case tar.TypeReg:
			// Create directory for file if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create directory for file '%s': %w", target, err)
			}

			// Create file
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file '%s': %w", target, err)
			}

			// Copy content from tar to file
			if _, err := io.Copy(f, tr); err != nil {
				if closeErr := f.Close(); closeErr != nil {
					fmt.Fprintf(os.Stderr, "Error closing file: %v\n", closeErr)
				}
				return fmt.Errorf("failed to write to file '%s': %w", target, err)
			}
			if err := f.Close(); err != nil {
				return fmt.Errorf("failed to close file '%s': %w", target, err)
			}

		case tar.TypeSymlink:
			// Create symlink
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("failed to create symlink '%s': %w", target, err)
			}

		default:
			// Skip other types
			fmt.Printf("Skipping unsupported file type for '%s'\n", header.Name)
		}
	}

	fmt.Printf("Bundle extracted to '%s'\n", u.Options.OutputDir)
	return nil
}
