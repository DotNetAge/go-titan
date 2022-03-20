package runtime

// Watcher 用于监视变量的变更
type Watcher interface {
	Next() (Result, error)
	Stop()
}

type Result struct {
	Action string
	Data   interface{}
}
