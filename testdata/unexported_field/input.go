package unexported_field

type Person struct {
	Name string
	age  int
}

func main() {
	_ = &Person{
		Name: "",
	}
}
