# fillstruct

A Go tool that automatically fills missing fields in struct literals with their zero values.

## Installation

```bash
go install github.com/nametake/fillstruct/cmd/fillstruct@latest
```

## Usage

```bash
fillstruct --type <importpath.TypeName> [--type <importpath.TypeName>...] [pattern]
```

### Options

- `--type`: Target type in the format `importpath.TypeName` (required, can be specified multiple times)
- `[pattern]`: Package pattern to process (default: `./...`)

## Example

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
fillstruct --type github.com/example/myapp.Person ./...
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

## Features

- Fills missing fields with zero values:
  - `string` -> `""`
  - `int`, `float`, etc. -> `0`
  - `bool` -> `false`
  - `pointer`, `slice`, `map`, `interface` -> `nil`
  - `struct` -> `StructType{}`
- Supports multiple target types
- Preserves code formatting and comments
- Skips position-based literals (e.g., `Person{"Alice", 25}`)
- Skips unexported fields when the struct is from another package

## License

MIT
