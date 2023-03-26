package main

import "strconv"

type StringFunction func(string)

type Interface interface {
	Method_A() string
	Method_B(int) (string, error)
	Method_C(func(int) error) (int, error)
}
type Object struct {
}

func (o *Object) Method_A() string {
	return ""
}
func (o *Object) Method_B(arg int) (string, error) {
	return strconv.Itoa(arg), nil
}
func (o *Object) Method_C(func(int) error) (int, error) {
	return 999, nil
}

func Function(p Interface, f StringFunction) {
	f(p.Method_A())
}

func main() {

}
