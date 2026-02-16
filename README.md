# fillstruct

A Go tool that automatically fills missing fields in struct literals with their zero values or custom default values.

## Installation

```bash
go install github.com/nametake/fillstruct/cmd/fillstruct@latest
```

## Usage

```bash
go run github.com/nametake/fillstruct/cmd/fillstruct@latest \
  --type <importpath.TypeName> \
  [--default <TypeSpec=ConstantName>...] \
  [pattern]
```

### Options

- `--type`: Target type in the format `importpath.TypeName` (required, can be specified multiple times)
- `--default`: Custom default value in the format `TypeSpec=ConstantName` (optional, can be specified multiple times)
  - For named types: `importpath.TypeName=ConstantName` (e.g., `github.com/example.Status=StatusUnknown`)
  - For basic types: `TypeName=Value` (e.g., `int=8080`, `bool=true`)
- `[pattern]`: Package pattern to process (default: `./...`)

## Examples

### Basic Usage

Given the following code:

```go
package main

type Person struct {
    Name        string
    Age         int
    Description *string
}

func main() {
    p := &Person{
        Name: "Alice",
    }
    _ = p
}
```

Run fillstruct:

```bash
go run github.com/nametake/fillstruct/cmd/fillstruct@latest \
  --type github.com/example/myapp.Person \
  ./...
```

The code will be updated to:

```go
package main

type Person struct {
    Name        string
    Age         int
    Description *string
}

func main() {
    p := &Person{
        Name:        "Alice",
        Age:         0,
        Description: nil,
    }
    _ = p
}
```

### Custom Default Values

You can specify custom default values for specific types:

```go
package main

type Status int

const (
    StatusUnknown Status = 0
    StatusActive  Status = 1
)

type Config struct {
    Name   string
    Port   int
    Status Status
}

func main() {
    c := &Config{
        Name: "myapp",
    }
    _ = c
}
```

Run fillstruct with custom defaults:

```bash
go run github.com/nametake/fillstruct/cmd/fillstruct@latest \
  --type github.com/example/myapp.Config \
  --default 'github.com/example/myapp.Status=StatusUnknown' \
  --default 'int=8080' \
  ./...
```

The code will be updated to:

```go
package main

type Status int

const (
    StatusUnknown Status = 0
    StatusActive  Status = 1
)

type Config struct {
    Name   string
    Port   int
    Status Status
}

func main() {
    c := &Config{
        Name:   "myapp",
        Port:   8080,
        Status: StatusUnknown,
    }
    _ = c
}
```

## Features

- Fills missing fields with zero values or custom default values:
  - `string` -> `""` (or custom default)
  - `int`, `float`, etc. -> `0` (or custom default)
  - `bool` -> `false` (or custom default)
  - `pointer`, `slice`, `map`, `interface` -> `nil`
  - `struct` -> `StructType{}`
  - Custom types -> Custom default constant (e.g., `StatusUnknown`)
- Supports custom default values for:
  - Named types (e.g., `type Status int`)
  - Basic types (e.g., `int`, `string`, `bool`)
- Supports multiple target types
- Preserves code formatting and comments
- Skips position-based literals (e.g., `Person{"Alice", 25}`)
- Skips unexported fields when the struct is from another package

## License

MIT
