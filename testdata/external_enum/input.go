package external_enum

import "github.com/nametake/fillstruct/testdata/external_enum/otherpkg"

type Config struct {
	Name   string
	Status otherpkg.Status
}

func main() {
	_ = &Config{
		Name: "test",
	}
}
