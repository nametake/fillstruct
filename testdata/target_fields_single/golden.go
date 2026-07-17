package target_fields_single

type User struct {
	Name  string
	Age   int
	Email string
}

func main() {
	_ = &User{
		Name: "",
		Age:  0,
	}
}
