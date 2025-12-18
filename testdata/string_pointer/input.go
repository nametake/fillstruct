package string_pointer

type Person struct {
	Name        string
	Age         int
	Description *string
}

func main() {
	_ = &Person{
		Name: "",
		Age:  0,
	}
}
