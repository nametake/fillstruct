package typewithpackage

import "time"

type Person struct {
	Name      string
	Age       int
	CreatedAt time.Time
}

func main() {
	_ = &Person{
		Name:      "",
		Age:       0,
		CreatedAt: time.Time{},
	}
}
