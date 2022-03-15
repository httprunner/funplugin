package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/rs/zerolog/log"
)

// CallFunc calls function with arguments
func CallFunc(fn reflect.Value, args ...interface{}) (interface{}, error) {
	argumentsValue, err := convertArgs(fn, args...)
	if err != nil {
		log.Error().Err(err).Msg("convert arguments failed")
		return nil, err
	}
	return call(fn, argumentsValue)
}

func convertArgs(fn reflect.Value, args ...interface{}) ([]reflect.Value, error) {
	fnArgsNum := fn.Type().NumIn()

	// function arguments should match exactly if function's last argument is not slice
	if len(args) != fnArgsNum && (fnArgsNum == 0 || fn.Type().In(fnArgsNum-1).Kind() != reflect.Slice) {
		return nil, fmt.Errorf("function expect %d arguments, but got %d", fnArgsNum, len(args))
	}

	argumentsValue := make([]reflect.Value, len(args))
	for index := 0; index < len(args); index++ {
		argument := args[index]
		if argument == nil {
			argumentsValue[index] = reflect.Zero(fn.Type().In(index))
			continue
		}

		argumentValue := reflect.ValueOf(argument)
		actualArgumentType := reflect.TypeOf(argument)

		var expectArgumentType reflect.Type
		if (index == fnArgsNum-1 && fn.Type().In(fnArgsNum-1).Kind() == reflect.Slice) || index > fnArgsNum-1 {
			// last fn argument is slice
			expectArgumentType = fn.Type().In(fnArgsNum - 1).Elem() // slice element type

			// last argument is also slice, e.g. []int
			if actualArgumentType.Kind() == reflect.Slice {
				if actualArgumentType.Elem() != expectArgumentType {
					err := fmt.Errorf("function argument %d's slice element type is not match, expect %v, actual %v",
						index, expectArgumentType, actualArgumentType)
					return nil, err
				}
				argumentsValue[index] = argumentValue
				continue
			}
		} else {
			expectArgumentType = fn.Type().In(index)
		}

		// type match
		if expectArgumentType == actualArgumentType {
			argumentsValue[index] = argumentValue
			continue
		}

		// type not match, check if convertible
		if !actualArgumentType.ConvertibleTo(expectArgumentType) {
			// function argument type not match and not convertible
			err := fmt.Errorf("function argument %d's type is neither match nor convertible, expect %v, actual %v",
				index, expectArgumentType, actualArgumentType)
			return nil, err
		}
		// convert argument to expect type
		argumentsValue[index] = argumentValue.Convert(expectArgumentType)
	}
	return argumentsValue, nil
}

func call(fn reflect.Value, args []reflect.Value) (interface{}, error) {
	resultValues := fn.Call(args)
	if resultValues == nil {
		// no returns
		return nil, nil
	} else if len(resultValues) == 2 {
		// return two arguments: interface{}, error
		if resultValues[1].Interface() != nil {
			return resultValues[0].Interface(), resultValues[1].Interface().(error)
		} else {
			return resultValues[0].Interface(), nil
		}
	} else if len(resultValues) == 1 {
		// return one argument
		if err, ok := resultValues[0].Interface().(error); ok {
			// return error
			return nil, err
		} else {
			// return interface{}
			return resultValues[0].Interface(), nil
		}
	} else {
		// return more than 2 arguments, unexpected
		err := fmt.Errorf("function should return at most 2 values")
		return nil, err
	}
}

// LocateFile searches destFile upward recursively until current
// working directory or system root dir.
func LocateFile(startPath string, destFile string) (string, error) {
	stat, err := os.Stat(startPath)
	if os.IsNotExist(err) {
		return "", err
	}

	var startDir string
	if stat.IsDir() {
		startDir = startPath
	} else {
		startDir = filepath.Dir(startPath)
	}
	startDir, _ = filepath.Abs(startDir)

	// convention over configuration
	pluginPath := filepath.Join(startDir, destFile)
	if _, err := os.Stat(pluginPath); err == nil {
		return pluginPath, nil
	}

	// current working directory
	cwd, _ := os.Getwd()
	if startDir == cwd {
		return "", fmt.Errorf("searched to CWD, plugin file not found")
	}

	// system root dir
	parentDir, _ := filepath.Abs(filepath.Dir(startDir))
	if parentDir == startDir {
		return "", fmt.Errorf("searched to system root dir, plugin file not found")
	}

	return LocateFile(parentDir, destFile)
}
