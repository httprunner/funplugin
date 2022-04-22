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
func EnsurePython3Venv(venvDir string, packages ...string) (python3 string, err error) {
	if runtime.GOOS == "windows" {
		python3 = filepath.Join(venvDir, "Scripts", "python.exe")
	} else {
		python3 = filepath.Join(venvDir, "bin", "python")
	}

	log.Info().
		Str("python", python3).
		Strs("packages", packages).
		Msg("ensure python3 venv")

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

	for _, pkg := range packages {
		err := InstallPythonPackage(python3, pkg)
		if err != nil {
			return python3, errors.Wrap(err, fmt.Sprintf("pip install %s failed", pkg))
		}
	}

	return python3, nil
}

func InstallPythonPackage(python3 string, pkg string) (err error) {
	var pkgName string
	if strings.Contains(pkg, "==") {
		// funppy==0.4.2
		pkgInfo := strings.Split(pkg, "==")
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

	// check if funppy installed
	err = execCommand(python3, "-m", "pip", "show", pkgName, "--quiet")
	if err == nil {
		// package is installed
		return nil
	}

	log.Info().Str("package", pkg).Msg("installing python package")

	// install package
	err = execCommand(python3, "-m",
		"pip", "install", pkg,
		"--quiet", "--disable-pip-version-check")
	if err != nil {
		return errors.Wrap(err, "pip install package failed")
	}

	return nil
}

func execCommand(cmdName string, args ...string) error {
	cmd := exec.Command(cmdName, args...)
	log.Info().Str("cmd", cmd.String()).Msg("exec command")

	// print output with colors
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Error().Err(err).Msg("exec command failed")
		return err
	}

	return nil
}
