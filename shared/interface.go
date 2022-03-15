package shared

// IFuncCaller is the interface that we're exposing as a plugin.
type IFuncCaller interface {
	GetNames() ([]string, error)                                    // get all plugin function names list
	Call(funcName string, args ...interface{}) (interface{}, error) // call plugin function
}
