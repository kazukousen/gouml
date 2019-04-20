package example

type Number int

const (
	One   Number = 1
	Two   Number = 2
	Three Number = 2
)

// not convertable (literal)
const (
	Un   = 1
	dos  = 2
	tres = 3
)
