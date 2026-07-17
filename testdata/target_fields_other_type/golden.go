package target_fields_other_type

type Config struct {
	Host  string
	Port  int
	Debug bool
}

type Person struct {
	Name string
	Age  int
}

func main() {
	_ = &Config{
		Host: "",
		Port: 0,
	}
	_ = &Person{
		Name: "",
		Age:  0,
	}
}
