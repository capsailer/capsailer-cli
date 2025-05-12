# Command Reference

Capsailer provides a set of commands for deploying applications in air-gapped environments.

## Available Commands

| Command | Description |
|---------|-------------|
| [`init`](init.md) | Validate and normalize the manifest file |
| [`build`](build.md) | Download and package images and charts into a bundle |
| [`unpack`](unpack.md) | Extract bundle and set up local registry |
| [`registry`](registry.md) | Deploy a Docker registry in a Kubernetes cluster |
| [`push`](push.md) | Push container images and charts to the registry |

## Workflow

1. **Connected Environment**
   - `init` - Validate your manifest file
   - `build` - Create a bundle with all required artifacts

2. **Air-Gapped Environment**
   - `registry` - Set up a local registry
   - `push` - Upload artifacts from the bundle
   - Use standard Helm commands to deploy applications (see [Working with Deployed Charts](deploy.md))

## Global Flags

Some flags are available across multiple commands:

| Flag | Description | Commands |
|------|-------------|----------|
| `--kubeconfig` | Path to kubeconfig file | `registry`, `push` |
| `--namespace` | Kubernetes namespace to use | `registry`, `push` |

## Getting Help

To get help for any command, use the `--help` flag:

```bash
capsailer --help
capsailer registry --help
``` 