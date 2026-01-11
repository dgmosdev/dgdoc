# dgdoc CI/CD Pipeline

This directory contains GitHub Actions workflows for automated testing, building, and releasing.

## Workflows

### CI Workflow (`ci.yml`)

Runs on every push and pull request to `main` and `develop` branches.

**Jobs:**
- **Test**: Runs tests on multiple OS (Ubuntu, macOS, Windows) and Go versions (1.21, 1.22)
  - Executes unit tests with race detection
  - Generates code coverage
  - Runs benchmarks
  - Uploads coverage to Codecov

- **Lint**: Runs golangci-lint for code quality checks

- **Build**: Builds the CLI and library on all supported platforms

**Matrix Testing:**
- Operating Systems: Ubuntu, macOS, Windows
- Go Versions: 1.21, 1.22

### Release Workflow (`release.yml`)

Triggers on version tags (e.g., `v1.0.0`).

**Features:**
- Automated binary builds for multiple platforms
- GitHub release creation
- Asset uploading
- Uses GoReleaser for cross-platform builds

## Local Testing

Run the same checks locally:

```bash
# Run tests
go test -v -race -coverprofile=coverage.txt ./...

# Run benchmarks
go test -bench=. -benchmem ./benchmarks

# Run linter (requires golangci-lint)
golangci-lint run --timeout=5m

# Build
go build ./cmd/dgdoc
```

## Coverage

Code coverage reports are automatically uploaded to Codecov on successful CI runs from Ubuntu with Go 1.22.

View coverage: `https://codecov.io/gh/dgmosdev/dgdoc`

## Release Process

1. Create and push a version tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. GitHub Actions automatically:
   - Runs all tests
   - Builds binaries for multiple platforms
   - Creates a GitHub release
   - Uploads release assets

## Status Badges

Add to README.md:

```markdown
[![CI](https://github.com/dgmosdev/dgdoc/workflows/CI/badge.svg)](https://github.com/dgmosdev/dgdoc/actions)
[![codecov](https://codecov.io/gh/dgmosdev/dgdoc/branch/main/graph/badge.svg)](https://codecov.io/gh/dgmosdev/dgdoc)
[![Go Report Card](https://goreportcard.com/badge/github.com/dgmosdev/dgdoc)](https://goreportcard.com/report/github.com/dgmosdev/dgdoc)
```
