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

Before making code changes:
- Ask if there is an existing GitHub issue for the change
- If no issue exists, use the gh CLI to create one:
  ```
  gh issue create
    --title "type: descriptive title"
    --body "detailed description"
    --label "type" (bug, enhancement, etc.)
  ```
- Include in issue description:
  - Current behavior (if applicable)
  - Expected behavior
  - Implementation approach
  - Testing steps
  - Related files
- Use issue discussions to track implementation decisions
- Reference the issue number in commit messages

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

When referencing issues in commits:
- For fixes, use "fixes #X" or "resolves #X"
- For related changes, use "relates to #X"
- Place issue references on a new line in the commit message body
- When amending commits with issue references, ensure the reference is preserved

Example commit message with issue reference:
```
feat: add new feature description

- Implementation detail 1
- Implementation detail 2

Resolves #42
```

Example issue creation command:
```
gh issue create --title "feat: add new capability" \
  --body "## Feature Description
- Current state
- Proposed changes
- Implementation approach

## Testing
- Step 1
- Step 2

## Related Files
- file1.go
- file2.go" \
  --label "enhancement"
```

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

