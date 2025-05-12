# Contributing to Capsailer

We welcome contributions to Capsailer! This document provides guidelines and instructions for contributing to the project.

## Getting Started

### Prerequisites

- Go 1.20 or later
- Git
- A Kubernetes cluster for testing (can be a local one like KinD or Minikube)

### Setting Up the Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/capsailer.git
   cd capsailer
   ```
3. Add the upstream repository as a remote:
   ```bash
   git remote add upstream https://github.com/jlnhnng/capsailer.git
   ```
4. Create a branch for your work:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Building and Testing

### Building the Project

```bash
# Build the binary
go build -o capsailer cmd/capsailer/main.go cmd/capsailer/commands.go

# Run the binary
./capsailer --help
```

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests for a specific package
go test -v ./pkg/deploy
```

### Integration Testing

For integration testing, you can use a local Kubernetes cluster:

```bash
# Create a KinD cluster
kind create cluster

# Build and run Capsailer
go build -o capsailer cmd/capsailer/main.go cmd/capsailer/commands.go
./capsailer registry
./capsailer push --image nginx:latest
./capsailer deploy --chart nginx
```

## Code Style and Guidelines

### Go Code Style

- Follow standard Go code style and conventions
- Use `gofmt` to format your code
- Run `golangci-lint` to check for issues:
  ```bash
  make lint
  ```

### Commit Messages

- Use clear, descriptive commit messages
- Start with a short summary line (50 chars or less)
- Optionally followed by a blank line and a more detailed explanation
- Reference issues and pull requests where appropriate

Example:
```
Add auto-discovery for charts in ChartMuseum

This change allows the deploy command to automatically find charts
in the ChartMuseum repository if they are not found locally.
It implements port-forwarding to access ChartMuseum from the CLI.

Fixes #123
```

## Pull Request Process

1. Update the documentation with details of changes to the interface, if applicable
2. Update the README.md or documentation with details of changes to the behavior
3. The PR should work for all supported platforms
4. Ensure all tests pass
5. Get at least one code review from a maintainer

## Adding New Features

When adding new features:

1. Start by opening an issue to discuss the feature
2. Write tests for your feature
3. Ensure the feature is well-documented
4. Keep the scope focused and specific

## Documentation

### Updating Documentation

The documentation is built using MkDocs with the Material theme:

1. Install the required tools:
   ```bash
   pip install mkdocs mkdocs-material mkdocs-minify-plugin
   ```

2. Make changes to the documentation in the `docs/` directory

3. Preview your changes:
   ```bash
   mkdocs serve
   ```

4. Build the documentation:
   ```bash
   mkdocs build
   ```

## Reporting Bugs

When reporting bugs:

1. Use the bug report template
2. Include detailed steps to reproduce the issue
3. Include information about your environment (OS, Go version, etc.)
4. Include logs or error messages if available

## Feature Requests

When requesting features:

1. Use the feature request template
2. Clearly describe the problem the feature would solve
3. Suggest a possible implementation if you have ideas

## Code of Conduct

Please note that this project adheres to a Code of Conduct. By participating, you are expected to uphold this code.

## Questions?

If you have any questions about contributing, feel free to open an issue or reach out to the maintainers. 