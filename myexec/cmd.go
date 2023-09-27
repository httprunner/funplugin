package myexec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/httprunner/funplugin/fungo"
	"github.com/pkg/errors"
)

var (
	logger         = fungo.Logger
	PYPI_INDEX_URL = os.Getenv("PYPI_INDEX_URL")
	PATH           = os.Getenv("PATH")
)

var python3Executable string = "python3" // system default python3

func isPython3(python string) bool {
	out, err := Command(python, "--version").Output()
	if err != nil {
		return false
	}
	if strings.HasPrefix(string(out), "Python 3") {
		return true
	}
	return false
}

// EnsurePython3Venv ensures python3 venv with specified packages
// venv should be directory path of target venv
func EnsurePython3Venv(venv string, packages ...string) (python3 string, err error) {
	// priority: specified > $HOME/.hrp/venv
	if venv == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "get user home dir failed")
		}
		venv = filepath.Join(home, ".hrp", "venv")
	}
	python3, err = ensurePython3Venv(venv, packages...)
	if err != nil {
		return "", err
	}
	python3Executable = python3
	logger.Info("set python3 executable path",
		"Python3Executable", python3Executable)
	return python3, nil
}

func ExecPython3Command(cmdName string, args ...string) error {
	args = append([]string{"-m", cmdName}, args...)
	return RunCommand(python3Executable, args...)
}

func AssertPythonPackage(python3 string, pkgName, pkgVersion string) error {
	out, err := Command(
		python3, "-c", fmt.Sprintf("\"import %s; print(%s.__version__)\"", pkgName, pkgName),
	).Output()
	if err != nil {
		return fmt.Errorf("python package %s not found", pkgName)
	}

	// do not check version if pkgVersion is empty
	if pkgVersion == "" {
		logger.Info("python package is ready", "name", pkgName)
		return nil
	}

	// check package version equality
	version := strings.TrimSpace(string(out))
	if strings.TrimLeft(version, "v") != strings.TrimLeft(pkgVersion, "v") {
		return fmt.Errorf("python package %s version %s not matched, please upgrade to %s",
			pkgName, version, pkgVersion)
	}

	logger.Info("python package is ready", "name", pkgName, "version", pkgVersion)
	return nil
}

func InstallPythonPackage(python3 string, pkg string) (err error) {
	var pkgName, pkgVersion string
	if strings.Contains(pkg, "==") {
		// specify package version
		// funppy==0.5.0
		pkgInfo := strings.Split(pkg, "==")
		pkgName = pkgInfo[0]
		pkgVersion = pkgInfo[1]
	} else {
		// package version not specified, install the latest by default
		// funppy
		pkgName = pkg
	}

	// check if package installed and version matched
	err = AssertPythonPackage(python3, pkgName, pkgVersion)
	if err == nil {
		return nil
	}

	// check if pip available
	err = RunCommand(python3, "-m", "pip", "--version")
	if err != nil {
		logger.Warn("pip is not available")
		return errors.Wrap(err, "pip is not available")
	}

	logger.Info("installing python package", "pkgName",
		pkgName, "pkgVersion", pkgVersion)

	// install package
	pypiIndexURL := PYPI_INDEX_URL
	if pypiIndexURL == "" {
		pypiIndexURL = "https://pypi.org/simple" // default
	}
	err = RunCommand(python3, "-m", "pip", "install", pkg, "--upgrade",
		"--index-url", pypiIndexURL,
		"--quiet", "--disable-pip-version-check")
	if err != nil {
		return errors.Wrap(err, "pip install package failed")
	}

	return AssertPythonPackage(python3, pkgName, pkgVersion)
}

func RunCommand(cmdName string, args ...string) error {
	cmd := Command(cmdName, args...)
	logger.Info("exec command", "cmd", cmd.String())

	// add cmd dir path to $PATH
	if cmdDir := filepath.Dir(cmdName); cmdDir != "" {
		path := fmt.Sprintf("%s:%s", cmdDir, PATH)
		if err := os.Setenv("PATH", path); err != nil {
			logger.Error("set env $PATH failed", "error", err)
			return err
		}
	}

	// print stderr output
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		logger.Error("exec command failed",
			"error", err, "stderr", stderrStr)
		if stderrStr != "" {
			err = errors.Wrap(err, stderrStr)
		}
		return err
	}

	return nil
}

func ExecCommandInDir(cmd *exec.Cmd, dir string) error {
	logger.Info("exec command", "cmd", cmd.String(), "dir", dir)
	cmd.Dir = dir

	// print stderr output
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		logger.Error("exec command failed",
			"error", err, "stderr", stderrStr)
		if stderrStr != "" {
			err = errors.Wrap(err, stderrStr)
		}
		return err
	}

	return nil
}
