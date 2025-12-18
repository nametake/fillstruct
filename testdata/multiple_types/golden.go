package multiple_types

type Person struct {
	Name string
	Age  int
}

type Company struct {
	Name    string
	Address string
}

func main() {
	_ = &Person{
		Name: "",
		Age:  0,
	}
	_ = &Company{
		Name:    "",
		Address: "",
	}
}
