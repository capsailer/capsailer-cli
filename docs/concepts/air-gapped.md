# Air-Gapped Environments

Air-gapped environments are isolated networks that are physically separated from the internet.

## Challenges

Deploying applications to Kubernetes in air-gapped environments presents several challenges:

1. **No access to public container registries**
2. **No access to public Helm repositories**
3. **Limited tooling availability**

## How Capsailer Helps

Capsailer provides an end-to-end solution:

1. **Bundle Creation** - Package all required images and charts
2. **Local Infrastructure** - Deploy a local registry and chart repository
3. **Self-Contained Tooling** - No dependency on external tools

## Best Practices

1. **Include the registry image** in your bundle
2. **Pin specific versions** of images and charts
3. **Verify bundle contents** before transfer 