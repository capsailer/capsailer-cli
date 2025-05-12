# Contributing

Contributions to Capsailer are welcome! This page provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Set up the development environment
4. Make your changes
5. Submit a pull request

## Development Environment

To set up a development environment:

```bash
# Clone the repository
git clone https://github.com/jlnhnng/capsailer.git
cd capsailer

# Install dependencies
go mod download

# Build the binary
go build -o capsailer cmd/capsailer/main.go
```

## Code Style

Please follow these code style guidelines:

- Use `gofmt` to format your code
- Write comments for exported functions and types
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Add tests for new functionality

## Testing

Before submitting a pull request, make sure your changes pass all tests:

```bash
go test ./...
```

## Pull Requests

When submitting a pull request:

1. Make sure your code passes all tests
2. Update documentation if necessary
3. Add a clear description of the changes
4. Reference any related issues

## Issues

When reporting issues, please include:

- A clear description of the problem
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Version information (Go version, OS, etc.)

## Feature Requests

Feature requests are welcome. Please provide:

- A clear description of the feature
- Use cases for the feature
- Any relevant examples or mockups

## License

By contributing to Capsailer, you agree that your contributions will be licensed under the project's [MIT License](https://github.com/jlnhnng/capsailer/blob/main/LICENSE). 