package shared

import (
	"fmt"
	"os/exec"
	"reflect"
	"strings"

	"github.com/pkg/errors"
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

func InstallPythonPackage(python3 string, pkg string) (err error) {
	var pkgName string
	if strings.Contains(pkg, "==") {
		// funppy==0.4.2
		pkgInfo := strings.Split(pkg, "==")
		pkgName = pkgInfo[0]
	} else if strings.Contains(pkg, ">=") {
		// httprunner>=4.0.0-beta
		pkgInfo := strings.Split(pkg, ">=")
		pkgName = pkgInfo[0]
	} else {
		pkgName = pkg
	}

	defer func() {
		if err == nil {
			// check package version
			if out, err := exec.Command(
				python3, "-c", fmt.Sprintf("import %s; print(%s.__version__)", pkgName, pkgName),
			).Output(); err == nil {
				log.Info().
					Str("name", pkgName).
					Str("version", strings.TrimSpace(string(out))).
					Msg("python package is ready")
			}
		}
	}()

	// check if pip available
	err = execCommand(python3, "-m", "pip", "--version")
	if err != nil {
		log.Warn().Msg("pip is not available")
		return errors.Wrap(err, "pip is not available")
	}

	// check if funppy installed
	err = execCommand(python3, "-m", "pip", "show", pkgName, "--quiet")
	if err == nil {
		// package is installed
		return nil
	}

	log.Info().Str("package", pkg).Msg("installing python package")

	// install package
	err = execCommand(python3, "-m", "pip", "install", "--upgrade", pkg,
		"--quiet", "--disable-pip-version-check")
	if err != nil {
		return errors.Wrap(err, "pip install package failed")
	}

	return nil
}

// ConvertCommonName returns name which deleted "_" and converted capital letter to their lower case
func ConvertCommonName(name string) string {
	return strings.ToLower(strings.Replace(name, "_", "", -1))
}
