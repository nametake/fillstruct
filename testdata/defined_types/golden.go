package defined_types

type Description string

type Person struct {
	Name        string
	Age         int
	Description Description
}

func main() {
	_ = &Person{
		Name:        "",
		Age:         0,
		Description: "",
	}
}
