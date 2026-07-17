# fillstruct

A Go tool that automatically fills missing fields in struct literals with their zero values or custom default values.

## Installation

```bash
go install github.com/nametake/fillstruct/cmd/fillstruct@latest
```

## Usage

```bash
go run github.com/nametake/fillstruct/cmd/fillstruct@latest \
  --type <importpath.TypeName[.FieldName]> \
  [--default <TypeSpec=ConstantName>...] \
  [pattern]
```

### Options

- `--type`: Target type in the format `importpath.TypeName` or `importpath.TypeName.FieldName` (required, can be specified multiple times)
  - `importpath.TypeName` fills all missing fields of the type
  - `importpath.TypeName.FieldName` fills only the specified field of the type
  - To fill multiple specific fields of the same type, repeat `--type` (e.g., `--type pkg.Foo.Name --type pkg.Foo.Age`)
  - If the same type is specified both with and without a field name, the one without a field name takes precedence and all fields are filled
- `--default`: Custom default value in the format `TypeSpec=ConstantName` (optional, can be specified multiple times)
  - For named types in the same package: `importpath.TypeName=ConstantName` (e.g., `github.com/example.Status=StatusUnknown`)
  - For named types in external packages: `importpath.TypeName=pkg.ConstantName` (e.g., `github.com/example/otherpkg.Status=otherpkg.StatusUnknown`)
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

### Filling Specific Fields Only

You can fill only specific fields by appending the field name to `--type`:

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

Run fillstruct with a field name:

```bash
go run github.com/nametake/fillstruct/cmd/fillstruct@latest \
  --type github.com/example/myapp.Person.Age \
  ./...
```

Only the specified field is filled:

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
        Age:  0,
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

### External Package Enums

When using enums from external packages, the constant name must include the package qualifier:

```go
package main

import "github.com/example/myapp/status"

type Config struct {
    Name   string
    Status status.Status
}

func main() {
    c := &Config{
        Name: "myapp",
    }
    _ = c
}
```

Run fillstruct with external package enum:

```bash
go run github.com/nametake/fillstruct/cmd/fillstruct@latest \
  --type github.com/example/myapp.Config \
  --default 'github.com/example/myapp/status.Status=status.StatusUnknown' \
  ./...
```

The code will be updated to:

```go
package main

import "github.com/example/myapp/status"

type Config struct {
    Name   string
    Status status.Status
}

func main() {
    c := &Config{
        Name:   "myapp",
        Status: status.StatusUnknown,
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
- Supports filling only specific fields (e.g., `--type pkg.Foo.Name`)
- Preserves code formatting and comments
- Skips position-based literals (e.g., `Person{"Alice", 25}`)
- Skips unexported fields when the struct is from another package

## License

MIT
