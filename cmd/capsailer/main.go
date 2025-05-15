package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/capsailer/capsailer-cli/pkg/build"
	"github.com/capsailer/capsailer-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "capsailer",
	Short: "Capsailer - A tool for air-gapped Kubernetes deployments",
	Long: `Capsailer is a command-line tool for packaging and deploying
Kubernetes applications in air-gapped (offline) environments.

It handles the complete lifecycle:
1. Download container images and Helm charts
2. Package everything into a portable bundle
3. Deploy in an air-gapped environment with a local registry`,
	SilenceErrors: true,
	Version:       "0.2.0",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize and validate a manifest file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit(manifestFile)
	},
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a deployable bundle from a manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuild(manifestFile, outputFile, rewriteImageRefs, registryURL)
	},
}

var unpackCmd = &cobra.Command{
	Use:   "unpack",
	Short: "Unpack a bundle in an air-gapped environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUnpack(bundleFile)
	},
}

// Global flags
var manifestFile string
var outputFile string
var bundleFile string
var kubeconfigPath string
var registryNamespace string
var rewriteImageRefs bool
var registryURL string

func init() {
	// init command flags
	initCmd.Flags().StringVar(&manifestFile, "manifest", "manifest.yaml", "Path to the manifest file")

	// build command flags
	buildCmd.Flags().StringVar(&manifestFile, "manifest", "manifest.yaml", "Path to the manifest file")
	buildCmd.Flags().StringVar(&outputFile, "output", "capsailer-bundle.tar.gz", "Output file path")
	buildCmd.Flags().BoolVar(&rewriteImageRefs, "rewrite-image-references", false, "Rewrite image references in Helm charts to use a private registry")
	buildCmd.Flags().StringVar(&registryURL, "registry-url", "", "URL of the private registry to use when rewriting image references")

	// unpack command flags
	unpackCmd.Flags().StringVar(&bundleFile, "file", "", "Path to the bundle file")
	if err := unpackCmd.MarkFlagRequired("file"); err != nil {
		fmt.Printf("Error marking flag as required: %v\n", err)
	}

	// Add commands to root
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(unpackCmd)
	// registry and push commands added in commands.go
}

// runInit handles the init command
func runInit(manifestPath string) error {
	fmt.Printf("Initializing manifest from %s\n", manifestPath)

	// Load and validate the manifest
	manifest, err := utils.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Print summary
	fmt.Printf("Manifest is valid. Found %d images and %d charts.\n",
		len(manifest.Images), len(manifest.Charts))

	// If there are charts, provide information about image reference analysis
	if len(manifest.Charts) > 0 {
		fmt.Println("\nHelm Chart Image Reference Analysis:")
		fmt.Println("-----------------------------------")
		fmt.Println("During the build process, Capsailer can analyze Helm charts")
		fmt.Println("for container image references and rewrite them to use your private registry.")
		fmt.Println("")
		fmt.Println("To enable this feature, use the following flags with the build command:")
		fmt.Println("  --rewrite-image-references    Enable image reference rewriting")
		fmt.Println("  --registry-url <url>          URL of your private registry")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  capsailer build --manifest manifest.yaml --output bundle.tar.gz \\")
		fmt.Println("    --rewrite-image-references --registry-url registry.local:5000")
		
		// Check if there are potential image references in chart names
		var potentialImageRefs []string
		for _, chart := range manifest.Charts {
			// Common chart names that likely use images with the same name
			if chart.Name == "nginx" || chart.Name == "redis" || chart.Name == "postgresql" || 
			   chart.Name == "mysql" || chart.Name == "mongodb" || chart.Name == "elasticsearch" ||
			   chart.Name == "prometheus" || chart.Name == "grafana" {
				
				// Check if an image with this name is in the manifest
				found := false
				for _, img := range manifest.Images {
					if strings.Contains(img, chart.Name) {
						found = true
						break
					}
				}
				
				if !found {
					potentialImageRefs = append(potentialImageRefs, chart.Name)
				}
			}
		}
		
		// Warn about potential missing images
		if len(potentialImageRefs) > 0 {
			fmt.Println("\nPotential missing images:")
			fmt.Println("------------------------")
			fmt.Println("The following charts may require images that are not in your manifest:")
			for _, name := range potentialImageRefs {
				fmt.Printf("  - Chart '%s' may need image '%s' or 'bitnami/%s'\n", name, name, name)
			}
			fmt.Println("\nConsider adding these images to your manifest to ensure they're included in the bundle.")
		}
	}

	return nil
}

// runBuild handles the build command
func runBuild(manifestPath, outputPath string, rewriteImageRefs bool, registryURL string) error {
	fmt.Printf("Building bundle from manifest %s\n", manifestPath)

	// Create builder with options
	builder := build.NewBuilder(build.BuildOptions{
		ManifestPath:          manifestPath,
		OutputPath:            outputPath,
		Parallel:              4,
		RewriteImageReferences: rewriteImageRefs,
		RegistryURL:           registryURL,
	})

	// Run the build
	if err := builder.Build(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	return nil
}

// runUnpack handles the unpack command
func runUnpack(bundlePath string) error {
	fmt.Printf("Unpacking bundle from %s\n", bundlePath)

	// Create unpacker with options
	unpacker := utils.NewUnpacker(utils.UnpackOptions{
		BundlePath: bundlePath,
		OutputDir:  ".",
	})

	// Extract the bundle
	if err := unpacker.Unpack(); err != nil {
		return fmt.Errorf("unpacking failed: %w", err)
	}

	fmt.Println("Bundle unpacked successfully.")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
