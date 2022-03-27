package shared

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
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

// EnsurePython3Venv ensures python3 venv for hashicorp python plugin
// venvDir should be directory path of target venv
func EnsurePython3Venv(venvDir string) (python3 string, err error) {
	if runtime.GOOS == "windows" {
		python3 = filepath.Join(venvDir, "Scripts", "python.exe")
	} else {
		python3 = filepath.Join(venvDir, "bin", "python")
	}

	defer func() {
		if err == nil {
			out, _ := exec.Command(
				python3, "-c", "import funppy; print(funppy.__version__)",
			).Output()
			log.Info().
				Str("venvDir", venvDir).
				Str("funppyVersion", strings.TrimSpace(string(out))).
				Msg("python3 venv is ready")
		}
	}()

	// check if python3 venv is available
	if err := exec.Command(python3, "--version").Run(); err != nil {
		// python3 venv not available, create one
		// check if system python3 is available
		if err := execCommand("python3", "--version"); err != nil {
			return "", errors.Wrap(err, "python3 not found")
		}

		// check if .venv exists
		if _, err := os.Stat(venvDir); err == nil {
			// .venv exists, remove first
			if err := execCommand("rm", "-rf", venvDir); err != nil {
				return "", errors.Wrap(err, "remove existed venv failed")
			}
		}

		// create python3 .venv
		// notice: --symlinks should be specified for windows
		// https://github.com/actions/virtual-environments/issues/2690
		if err := execCommand("python3", "-m", "venv", "--symlinks", venvDir); err != nil {
			return "", errors.Wrap(err, "create python3 venv failed")
		}
	}

	// check if funppy installed
	err = exec.Command(python3, "-m", "pip", "show", "funppy", "--quiet").Run()
	if err == nil {
		// funppy is installed
		return python3, nil
	}

	// install funppy
	err = execCommand(python3, "-m",
		"pip", "install", "funppy", "--quiet", "--disable-pip-version-check")
	if err != nil {
		return "", errors.Wrap(err, "install funppy failed")
	}
	return python3, nil
}

func execCommand(cmdName string, args ...string) error {
	cmd := exec.Command(cmdName, args...)
	log.Info().Str("cmd", cmd.String()).Msg("exec command")
	output, err := cmd.CombinedOutput()
	out := strings.TrimSpace(string(output))
	if err != nil {
		log.Error().Err(err).Str("output", out).Msg("exec command failed")
	} else if len(out) != 0 {
		log.Info().Str("output", out).Msg("exec command success")
	}
	return err
}
