package main

import (
	"fmt"
	"os"

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
	Version: "0.1.0",
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
		return runBuild(manifestFile, outputFile)
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

func init() {
	// init command flags
	initCmd.Flags().StringVar(&manifestFile, "manifest", "manifest.yaml", "Path to the manifest file")
	
	// build command flags
	buildCmd.Flags().StringVar(&manifestFile, "manifest", "manifest.yaml", "Path to the manifest file")
	buildCmd.Flags().StringVar(&outputFile, "output", "capsailer-bundle.tar.gz", "Output file path")
	
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}