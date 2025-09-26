# OmniDrop Documentation

This directory contains comprehensive documentation for the OmniDrop project.

## Documentation Overview

| Document | Purpose | Audience |
|----------|---------|----------|
| [API Reference](api-reference.yaml) | OpenAPI 3.0 specification | Developers, integrators |
| [Architecture](architecture.md) | System design and data flow | Developers, architects |
| [Developer Guide](developer-guide.md) | Contributing and development | Contributors |
| [Deployment Guide](deployment-guide.md) | Production setup and operations | DevOps, administrators |

## Quick Navigation

### For API Users
- **Getting Started**: See main [README.md](../README.md) for quick setup
- **API Details**: Review [API Reference](api-reference.yaml) for complete endpoint documentation
- **Integration Examples**: Check [Developer Guide](developer-guide.md) for code samples

### For Developers
- **Code Structure**: Start with [Architecture](architecture.md) overview
- **Contributing**: Follow [Developer Guide](developer-guide.md) for setup and standards
- **Testing**: Use isolated environments as described in developer guide

### For Administrators
- **Production Setup**: Follow [Deployment Guide](deployment-guide.md) step by step
- **Monitoring**: Configure health checks and log rotation per deployment guide
- **Security**: Review security best practices in deployment documentation

## Documentation Standards

### Format
- **Markdown**: Primary format for human-readable docs
- **OpenAPI**: YAML format for API specification
- **ASCII Diagrams**: Simple text-based visualizations

### Structure
- **Clear headings**: Hierarchical organization
- **Code examples**: Practical, runnable snippets
- **Cross-references**: Links between related documents

### Maintenance
- **Version synchronization**: Keep docs in sync with code changes
- **Example validation**: Test all code examples before release
- **Regular review**: Update for accuracy and completeness

## Contributing to Documentation

### Adding New Documentation
1. Follow existing structure and naming conventions
2. Include practical examples and code snippets
3. Cross-reference related documents
4. Update this index file

### Updating Existing Docs
1. Maintain backward compatibility for API docs
2. Version significant changes
3. Update modification dates
4. Test all examples and links

## Tools and Validation

### OpenAPI Validation
```bash
# Install swagger-codegen for validation
npm install -g swagger-codegen-cli

# Validate API specification
swagger-codegen validate -i docs/api-reference.yaml
```

### Markdown Linting
```bash
# Install markdownlint
npm install -g markdownlint-cli

# Lint all markdown files
markdownlint docs/*.md
```

### Link Checking
```bash
# Install markdown-link-check
npm install -g markdown-link-check

# Check all links
find docs -name "*.md" -exec markdown-link-check {} \;
```

## Feedback and Improvements

Documentation improvements are welcome! Please:

1. **Open issues** for inaccuracies or missing information
2. **Submit pull requests** for corrections and enhancements
3. **Suggest structure improvements** for better navigation
4. **Report broken examples** or outdated information

---

For questions about documentation, please open an issue in the main repository.