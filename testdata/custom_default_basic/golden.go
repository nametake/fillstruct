package custom_default_basic

type Config struct {
	Name    string
	Port    int
	Enabled bool
}

func main() {
	_ = &Config{
		Name:    "test",
		Port:    8080,
		Enabled: true,
	}
}
