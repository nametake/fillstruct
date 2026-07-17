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
	}
	_ = &Person{
		Name: "",
	}
}
