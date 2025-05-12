# Capsailer

Capsailer is a CLI tool for delivering Kubernetes applications into air-gapped (offline) environments.

## Overview

Capsailer allows you to define Helm charts and container images, package them into a portable archive, and deploy them into an air-gapped Kubernetes environment by installing a local registry and Helm chart server.

## Features

- Download container images and Helm charts from public or private repositories
- Package everything into a single, portable archive file
- Deploy the bundle in an air-gapped environment
- Set up a local container registry and Helm chart repository
- Self-contained CLI that doesn't require Docker or skopeo for image operations
- Built-in support for pushing container images and Helm charts without external dependencies
- Automated chart repository deployment and publishing

## How It Works

Capsailer provides an all-in-one solution for air-gapped Kubernetes deployments:

1. **Bundle Creation**: Package container images and Helm charts into a portable bundle
2. **Registry Infrastructure**: Deploy a container registry and Helm chart repository
3. **Push Mechanism**: Upload images and charts without requiring external tools like Docker or skopeo

Unlike other solutions, Capsailer handles both container images and Helm charts natively, without relying on external tools in the air-gapped environment. 