package position_based

type Person struct {
	Name string
	Age  int
}

func main() {
	_ = &Person{"", 0}
}
