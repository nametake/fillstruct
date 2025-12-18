package test_file

import "testing"

type TestCase struct {
	Name  string
	Input int
	Want  int
}

func TestSomething(t *testing.T) {
	tc := &TestCase{
		Name:  "test",
		Input: 0,
		Want:  0,
	}
	_ = tc
}
