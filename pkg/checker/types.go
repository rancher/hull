package checker

type Check struct {
	Name string
	Func interface{}
}

type CheckFunc interface{}
