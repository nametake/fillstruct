package custom_default

type Status int

const (
	StatusUnknown  Status = 0
	StatusActive   Status = 1
	StatusInactive Status = 2
)

type Role int

const (
	RoleGuest Role = 0
	RoleUser  Role = 1
	RoleAdmin Role = 2
)

type Config struct {
	Name   string
	Status Status
	Role   Role
	Count  int
}

func main() {
	_ = &Config{
		Name:   "test",
		Status: StatusUnknown,
		Role:   RoleGuest,
		Count:  0,
	}
}
