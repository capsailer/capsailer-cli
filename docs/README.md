# Capsailer Documentation

This directory contains the documentation for Capsailer, built with [MkDocs](https://www.mkdocs.org/) and the [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) theme.

## Building the Documentation

To build the documentation:

1. Install the required dependencies:

```bash
pip install -r requirements.txt
```

2. Serve the documentation locally:

```bash
mkdocs serve
```

This will start a local server at http://localhost:8000 where you can preview the documentation.

3. Build the documentation:

```bash
mkdocs build
```

This will generate the static site in the `site` directory.

## Documentation Structure

- `index.md`: Home page
- `getting-started.md`: Quick start guide
- `user-guide/`: Detailed user guides
- `commands/`: Command reference
- `examples.md`: Usage examples
- `contributing.md`: Contribution guidelines 