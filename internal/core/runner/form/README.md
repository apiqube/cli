# Form Runner Package

This package provides a powerful and flexible system for processing forms and templates for ApiQube CLI.

## Architecture

The package is built on a modular architecture with clear separation of concerns:

### Main Components

#### 1. Runner (`runner_new.go`)
The main class that coordinates all form processing. It provides a simple API for:
- Processing strings with templates (`Apply`)
- Processing complex data structures (`ApplyBody`)
- Processing HTTP headers (`MapHeaders`)

#### 2. Processors (`processors.go`)
A system of processors for handling different data types:
- **StringProcessor**: Processes string values and templates
- **MapProcessor**: Processes objects (map[string]any)
- **ArrayProcessor**: Processes arrays
- **CompositeProcessor**: Combines all processors

#### 3. Template Resolver (`template_resolver.go`)
Responsible for resolving templates:
- Supports contextual variables
- Supports Body references (`Body.field.subfield`)
- Integrates with a template engine for functions

#### 4. Value Extractor (`value_extractor.go`)
Extracts values from nested data structures:
- Supports dot notation paths
- Supports array indices
- Supports dynamic indices (`#`)

#### 5. Directive Executor (`directive_executor.go`)
Executes special directives:
- Registers and executes directives
- Checks dependencies
- Extensible directive system

#### 6. Reference Resolver (`reference_resolver.go`)
Resolves references between data elements:
- Supports cyclic references
- Recursive structure processing
- Context-aware resolution

## Interfaces

All components implement well-defined interfaces (`interfaces.go`):

```go
type Processor interface {
    Process(ctx interfaces.ExecutionContext, value any, pass []*tests.Pass, processedData map[string]any, indexStack []int) any
}

type TemplateResolver interface {
    Resolve(ctx interfaces.ExecutionContext, template string, pass []*tests.Pass, processedData map[string]any, indexStack []int) (any, error)
}

type DirectiveExecutor interface {
    Execute(ctx interfaces.ExecutionContext, value any, pass []*tests.Pass, processedData map[string]any, indexStack []int) (any, error)
    CanHandle(value any) bool
}
// ... and others
```

## Usage

### Basic Usage

```go
runner := NewRunner()
ctx := // your ExecutionContext

// Process a string
result := runner.Apply(ctx, "Hello {{ username }}", nil)

// Process a complex structure
body := map[string]any{
    "user": map[string]any{
        "name": "{{ username }}",
        "email": "{{ Fake.email }}",
    },
}
processed := runner.ApplyBody(ctx, body, nil)

// Process headers
headers := map[string]string{
    "Authorization": "Bearer {{ token }}",
}
processedHeaders := runner.MapHeaders(ctx, headers, nil)
```

### Directives

Special directives are supported for advanced logic:

```go
body := map[string]any{
    "__repeat": 3,
    "__template": map[string]any{
        "id": "{{ Fake.int }}",
        "name": "User #{{ # }}",
    },
}
// Result: array of 3 objects with unique data
```

### Extending Functionality

```go
// Register a new directive
type MyDirective struct{}
func (d *MyDirective) Name() string { return "mydir" }
func (d *MyDirective) Dependencies() []string { return []string{} }
func (d *MyDirective) Execute(...) (any, error) { /* your logic */ }

runner.RegisterDirective(&MyDirective{})
```

## Refactoring Benefits

1. **Modularity**: Each component has a clear responsibility
2. **Testability**: Components are easy to test in isolation
3. **Extensibility**: Easy to add new processors and directives
4. **Readability**: Code is more understandable and structured
5. **Performance**: Optimized data processing
6. **Reliability**: Better error handling and edge case management

## Testing

The package includes a comprehensive test suite (`runner_test.go`) with mock objects for testing all components.

```bash
go test ./internal/core/runner/form/...
```

## Migration

To migrate from the old version:
1. Replace imports with new files
2. The API remains compatible
3. Additional features are available via new methods
