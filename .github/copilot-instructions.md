## CoPilot Instructions

Answer all questions in the style of a friendly colleague, using informal language.

Follow standard Go formatting and linting rules.

When suggesting code changes:
- Follow Go best practices
- Ensure code is properly formatted
- Include necessary error handling
- Add appropriate comments for complex logic

When writing Go documentation:
- Start package comments with "Package [name]" and describe the package's purpose
- Document exported names with complete sentences starting with the name
- Include examples in test files named example_*_test.go
- Structure examples as full working programs in Example[Function] format
- Include Output: comments in examples to enable automated testing
- Use doc.go for extensive package documentation
- Document any error types, constants, and variables
- Follow the standard "godoc" format for documentation

Example documentation format:
```go
// Package mypackage provides functionality for...
package mypackage

// MyType represents... and is used for...
type MyType struct {
    // Field descriptions if needed
}

// MyFunction does X and returns Y.
// It handles error case Z by...
func MyFunction() error {
    // Implementation
}
```

Example test format:
```go
func ExampleMyFunction() {
    result := MyFunction()
    fmt.Println(result)
    // Output: expected output
}
```

Use conventional commits format for all commit messages:
- feat: new features
- fix: bug fixes
- docs: documentation changes
- style: formatting changes
- refactor: code restructuring
- test: adding tests
- chore: maintenance tasks
- ci: CI/CD changes
- perf: performance improvements
- build: build system changes

