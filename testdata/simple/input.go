package simple

type Person struct {
	Name string
	Age  int
	Sex  string
}

func main() {
	_ = &Person{
		Name: "",
		Sex:  "",
	}
}
