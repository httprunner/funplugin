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

// PreparePython3Venv prepares python3 venv for hashicorp python plugin
// created .venv directory will be located besides the plugin file path
func PreparePython3Venv(path string) (python3 string, err error) {
	projectDir, _ := filepath.Abs(filepath.Dir(path))
	if err := ExecCommand(exec.Command("python3", "--version"), projectDir); err != nil {
		return "", errors.Wrap(err, "python3 not found")
	}

	venvDir := ".venv"
	if runtime.GOOS == "windows" {
		python3 = filepath.Join(projectDir, venvDir, "Scripts", "python3.exe")
	} else {
		python3 = filepath.Join(projectDir, venvDir, "bin", "python3")
	}

	// check if python .venv exists
	if !isExecutableFileExists(python3) {
		// create python .venv
		// notice: --symlinks should be specified for windows
		// https://github.com/actions/virtual-environments/issues/2690
		if err := ExecCommand(exec.Command("python3", "-m", "venv", "--symlinks", venvDir), projectDir); err != nil {
			return "", errors.Wrap(err, "create python3 venv failed")
		}
	}

	// check if funppy installed
	pip3CheckCmd := exec.Command(python3, "-m",
		"pip", "show", "funppy", "--quiet")
	if err := ExecCommand(pip3CheckCmd, projectDir); err == nil {
		// funppy is installed
		return python3, nil
	}

	// install funppy
	pip3InstallCmd := exec.Command(python3, "-m",
		"pip", "install", "funppy", "--quiet", "--disable-pip-version-check")
	if err := ExecCommand(pip3InstallCmd, projectDir); err != nil {
		return "", errors.Wrap(err, "install funppy failed")
	}
	return python3, nil
}

func ExecCommand(cmd *exec.Cmd, cwd string) error {
	log.Info().Str("cmd", cmd.String()).Str("cwd", cwd).Msg("exec command")
	cmd.Dir = cwd
	output, err := cmd.CombinedOutput()
	out := strings.TrimSpace(string(output))
	if err != nil {
		log.Error().Err(err).Str("output", out).Msg("exec command failed")
	} else if len(out) != 0 {
		log.Info().Str("output", out).Msg("exec command success")
	}
	return err
}

// isExecutableFileExists returns true if path exists and path is executable file
func isExecutableFileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// path not exists
		log.Warn().Str("path", path).Msg("path is not exists")
		return false
	}

	// path exists
	if !info.Mode().IsRegular() {
		// path is not regular file
		log.Warn().Str("path", path).Msg("path is not regular file")
		return false
	}

	// file path is executable
	if info.Mode().Perm()&0100 == 0 {
		log.Warn().Str("path", path).Msg("path is not executable")
		return false
	}

	return true
}
