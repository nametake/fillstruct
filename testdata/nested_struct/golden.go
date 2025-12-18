package nested_struct

type Address struct {
	City string
}

type Person struct {
	Name    string
	Address Address
}

func main() {
	_ = &Person{
		Name:    "",
		Address: Address{},
	}
}
