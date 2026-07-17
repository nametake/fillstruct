package target_fields_multiple

type User struct {
	Name    string
	Age     int
	Email   string
	Address string
}

func main() {
	_ = &User{
		Name:  "",
		Age:   0,
		Email: "",
	}
}
