package target_fields_unspecified

type User struct {
	Name  string
	Age   int
	Email string
}

func main() {
	_ = &User{
		Name:  "",
		Age:   0,
		Email: "",
	}
}
