package main

import "fmt"

// Bird 定义了一个鸟类
type Bird interface {
	Fly()
	Type() string
}

// Canary 鸟类：金丝雀
type Canary struct {
	Name string
}

func (c *Canary) Fly() {
	fmt.Printf("我是%s，用黄色的翅膀飞\n", c.Name)
}
func (c *Canary) Type() string {
	return c.Name
}

// Crow 鸟类：乌鸦
type Crow struct {
	Name string
}

func (c *Crow) Fly() {
	fmt.Printf("我是%s，用黑色的翅膀飞\n", c.Name)
}
func (c *Crow) Type() string {
	return c.Name
}

// LetItFly 让鸟类飞一下
func LetItFly(bird Bird) {
	fmt.Printf("Let %s Fly!\n", bird.Type())
	bird.Fly()
}

func main() {
	LetItFly(&Canary{"金丝雀"})
	LetItFly(&Crow{"乌鸦"})
}
