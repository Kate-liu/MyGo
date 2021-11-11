package main

import "fmt"

type Person struct {
	name string
	age  int
}

func NewPersonFactory(age int) func(name string) Person {
	return func(name string) Person {
		return Person{
			name: name,
			age:  age,
		}
	}
}

func main() {
	newBaby := NewPersonFactory(1)
	baby := newBaby("john")
	fmt.Println("baby is", baby.name, "age is", baby.age)

	newTeenager := NewPersonFactory(16)
	teen := newTeenager("jill")
	fmt.Println("teenager is", teen.name, "age is", teen.age)
}
