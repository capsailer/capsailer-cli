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

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a chart in an air-gapped environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDeploy(chartName, valuesFile)
	},
}

// Global flags
var manifestFile string
var outputFile string
var bundleFile string
var chartName string
var valuesFile string
var registryNamespace string
var registryImage string
var registryPersistent bool
var kubeconfigPath string

func init() {
	// init command flags
	initCmd.Flags().StringVar(&manifestFile, "manifest", "manifest.yaml", "Path to the manifest file")
	
	// build command flags
	buildCmd.Flags().StringVar(&manifestFile, "manifest", "manifest.yaml", "Path to the manifest file")
	buildCmd.Flags().StringVar(&outputFile, "output", "capsailer-bundle.tar.gz", "Output file path")
	
	// unpack command flags
	unpackCmd.Flags().StringVar(&bundleFile, "file", "", "Path to the bundle file")
	unpackCmd.MarkFlagRequired("file")
	
	// deploy command flags
	deployCmd.Flags().StringVar(&chartName, "chart", "", "Name of the chart to deploy")
	deployCmd.Flags().StringVar(&valuesFile, "values", "", "Values file for the chart")
	deployCmd.MarkFlagRequired("chart")
	
	// registry command flags (moved to commands.go)
	
	// Add commands to root
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(unpackCmd)
	rootCmd.AddCommand(deployCmd)
	// registry command added in commands.go
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}