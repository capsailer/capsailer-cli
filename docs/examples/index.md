# Examples

This section provides practical examples of using Capsailer in various scenarios.

## Available Examples

### [Basic Usage](basic.md)

A simple example showing how to deploy a basic Nginx web server in an air-gapped environment. This is a good starting point for new users.

### [Complete Workflow](workflow.md)

A comprehensive example demonstrating the complete workflow from manifest creation to deploying multiple applications in an air-gapped environment.

### [Custom Values](custom-values.md)

Learn how to customize your deployments using custom values files with Capsailer.

## Common Use Cases

Capsailer is designed to handle a variety of air-gapped deployment scenarios:

### Single Application Deployment

The simplest use case is deploying a single application:

```bash
# In connected environment
capsailer build --manifest app-manifest.yaml --output app-bundle.tar.gz

# In air-gapped environment
capsailer registry
capsailer push --bundle app-bundle.tar.gz
capsailer deploy --chart app-name
```

### Multi-Tier Application Stack

For more complex applications with multiple components:

```bash
# In connected environment
capsailer build --manifest full-stack-manifest.yaml --output full-stack-bundle.tar.gz

# In air-gapped environment
capsailer registry
capsailer push --bundle full-stack-bundle.tar.gz
capsailer deploy --chart database --namespace data
capsailer deploy --chart backend --namespace app
capsailer deploy --chart frontend --namespace web
```

### Incremental Updates

For updating existing deployments:

```bash
# In connected environment
capsailer build --manifest update-manifest.yaml --output update-bundle.tar.gz

# In air-gapped environment
capsailer push --bundle update-bundle.tar.gz
capsailer deploy --chart app-name
```

## Tips for Creating Effective Examples

When creating your own examples or workflows:

1. **Start Small**: Begin with a minimal viable bundle and add components incrementally
2. **Include Registry Image**: Always include `registry:2` in your manifest for truly air-gapped deployments
3. **Use Version Pinning**: Always specify exact versions for images and charts
4. **Namespace Organization**: Use namespaces to organize multi-application deployments
5. **Custom Values**: Leverage custom values files for environment-specific configurations

## Getting Help

If you encounter issues with these examples or have questions about specific use cases, please:

- Check the [Concepts](../concepts/index.md) section for deeper understanding
- Refer to the [Command Reference](../commands/index.md) for detailed command information
- Open an issue on the [GitHub repository](https://github.com/jlnhnng/capsailer/issues) for support 