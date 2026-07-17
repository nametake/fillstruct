package target_fields_already_present

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
