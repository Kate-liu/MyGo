package hungry_singleton

type singleton struct {
}

var ins *singleton = &singleton{}

func GetInsOr() *singleton {
	return ins
}
