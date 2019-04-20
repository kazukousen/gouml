package example

type Foo struct {
	Bar Bar
	Baz []Baz
}

type Bar int

type Baz string

type Hoge float64

type Fuga []byte

func (f Foo) UsesHoge(hoge Hoge) bool {
	return hoge > 0
}

func (f Foo) ReturnsFuga() Fuga {
	return Fuga([]byte("new fuga"))
}
