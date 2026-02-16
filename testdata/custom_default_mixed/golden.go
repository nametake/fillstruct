package custom_default_mixed

type Status int

const (
	StatusUnknown Status = 0
	StatusActive  Status = 1
)

type User struct {
	Name   string
	Status Status
	Age    int
	Email  string
}

func main() {
	_ = &User{
		Name:   "alice",
		Status: StatusUnknown,
		Age:    0,
		Email:  "",
	}
}
