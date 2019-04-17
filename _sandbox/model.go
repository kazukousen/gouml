package main

type Animal interface {
	Cry() string
}

type Cat struct {
	Name string
}

func (c Cat) Cry() string {
	return "にゃんにゃん"
}

type Dog struct {
	Name string
}

func (d Dog) Cry() string {
	return "わんわん"
}

type CryFunc func(an Animal) string
