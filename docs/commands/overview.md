# Command Reference

Capsailer provides several commands for managing container images and Helm charts in air-gapped environments.

## Available Commands

| Command | Description |
|---------|-------------|
| `init` | Validate and normalize the manifest |
| `build` | Download and package images and charts |
| `registry` | Deploy a standalone Docker registry in a Kubernetes cluster |
| `push` | Push container images to the registry |
| `unpack` | Extract bundle and set up local registry |

## Global Flags

These flags are available for all commands:

| Flag | Description |
|------|-------------|
| `--help` | Show help for the command |
| `--kubeconfig` | Path to the kubeconfig file |
| `--namespace` | Kubernetes namespace to use |
| `--verbose` | Enable verbose output |

## Command Usage

Each command has its own set of flags and arguments. You can see the available options for a command by running:

```bash
capsailer [command] --help
```

## Command Workflow

The typical workflow for using Capsailer commands is:

1. `init`: Validate your manifest file
2. `build`: Build a bundle from the manifest
3. `registry`: Deploy a registry in the air-gapped environment
4. `push`: Push artifacts from the bundle to the registry

For more details on each command, see the individual command reference pages. 