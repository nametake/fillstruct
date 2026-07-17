package target_fields_empty

type User struct {
	Name  string
	Age   int
	Email string
}

func main() {
	_ = &User{
		Name: "",
	}
}
