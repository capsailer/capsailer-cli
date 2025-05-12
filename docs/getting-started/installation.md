# Installation

This guide explains how to install Capsailer on your system.

## Prerequisites

Before installing Capsailer, ensure you have the following:

- **Go 1.20 or later** - Required for building from source
- **Git** - For cloning the repository
- **Kubernetes cluster** - For deploying applications (only needed in the air-gapped environment)
- **kubectl** - For interacting with the Kubernetes cluster

## Installation Methods

### Method 1: Build from Source

The recommended way to install Capsailer is to build it from source:

```bash
# Clone the repository
git clone https://github.com/jlnhnng/capsailer.git
cd capsailer

# Build the binary
go build -o capsailer cmd/capsailer/main.go cmd/capsailer/commands.go

# Make it executable
chmod +x capsailer

# Move to a directory in your PATH (optional)
sudo mv capsailer /usr/local/bin/
```

### Method 2: Download Pre-built Binary

Alternatively, you can download a pre-built binary from the releases page:

```bash
# Download the latest release for your platform
# Replace X.Y.Z with the version number and OS-ARCH with your operating system and architecture
curl -LO https://github.com/jlnhnng/capsailer/releases/download/vX.Y.Z/capsailer-X.Y.Z-OS-ARCH.tar.gz

# Extract the binary
tar -xzf capsailer-X.Y.Z-OS-ARCH.tar.gz

# Make it executable
chmod +x capsailer

# Move to a directory in your PATH (optional)
sudo mv capsailer /usr/local/bin/
```

## Verifying the Installation

To verify that Capsailer is installed correctly, run:

```bash
capsailer --version
```

This should display the version of Capsailer.

## Installing in Both Environments

Remember that Capsailer needs to be installed in both:

1. **Connected Environment** - Where you'll build bundles
2. **Air-Gapped Environment** - Where you'll deploy applications

In the air-gapped environment, you can either:

- Build from source (if Go is available)
- Transfer the pre-built binary from a connected environment
- Include the binary in your organization's approved software distribution process

## Next Steps

Now that you have Capsailer installed, you can:

1. Create a [manifest file](manifest.md) defining your application requirements
2. Follow the [quick start guide](quick-start.md) to deploy your first application 