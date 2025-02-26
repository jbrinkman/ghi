# Developer Guide

This guide helps you get started with developing the GitHub Info CLI tool.

## Prerequisites

- Go 1.21 or higher
- Git
- Make

## Development Setup

1. Clone the repository:
   ```sh
   git clone https://github.com/jbrinkman/ghi.git
   cd ghi
   ```

2. Install dependencies:
   ```sh
   go mod download
   ```

## Building

The project uses Make for building and testing. Here are the main make targets:

- `make build`: Builds the binary in the local bin directory
- `make install`: Installs the binary to your GOPATH bin directory
- `make test`: Runs all tests
- `make clean`: Removes build artifacts
- `make tag VERSION=X.X.X`: Creates and pushes a new version tag

For development builds:
```sh
make build
```

The binary will be created in the `bin` directory.

For installation:
```sh
make install
```

This installs the binary to your `$GOPATH/bin` directory.

## Version Information

When building the application, version information is injected at build time. The version information includes:
- Version number
- Git commit hash
- Build date

To build with a specific version:
```sh
make build VERSION=1.0.0
```

## Contributing

### Pull Request Guidelines

1. **Branch Naming**
   - Use conventional prefixes:
     - `feat/` for new features
     - `fix/` for bug fixes
     - `docs/` for documentation changes
     - `style/` for formatting changes
     - `refactor/` for code restructuring
     - `test/` for adding tests
     - `chore/` for maintenance tasks

2. **Commit Messages**
   Follow the conventional commits format:
   - `feat: add new feature`
   - `fix: resolve specific issue`
   - `docs: update documentation`
   - `style: format code`
   - `refactor: restructure code`
   - `test: add tests`
   - `chore: update dependencies`
   - `ci: update workflow`
   - `perf: improve performance`
   - `build: update build process`

3. **Code Standards**
   - Follow Go best practices and standard formatting
   - Include appropriate error handling
   - Add comments for complex logic
   - Write tests for new functionality
   - Update documentation for API changes

4. **Testing**
   - Ensure all tests pass: `make test`
   - Add new tests for new functionality
   - Include example tests where appropriate

5. **Documentation**
   - Update README.md if adding new features
   - Include godoc comments for exported functions
   - Add example tests for new commands

6. **PR Description**
   - Describe the changes
   - Reference any related issues
   - Include steps to test the changes
   - List any breaking changes

### Release Process

1. Update version number in relevant files
2. Create a new version tag:
   ```sh
   make tag VERSION=X.X.X
   ```
3. The GitHub Actions workflow will automatically:
   - Build and test the code
   - Create a GitHub release
   - Publish to pkg.go.dev