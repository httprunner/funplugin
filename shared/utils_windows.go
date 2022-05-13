// +build windows

package shared

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// EnsurePython3Venv ensures python3 venv for hashicorp python plugin
// venvDir should be directory path of target venv
func EnsurePython3Venv(venvDir string, packages ...string) (python3 string, err error) {
	python3 = filepath.Join(venvDir, "Scripts", "python3.exe")
	log.Info().
		Str("python3", python3).
		Strs("packages", packages).
		Msg("ensure python3 venv")

	// check if python3 venv is available
	if err := exec.Command("cmd", "/c", python3, "--version").Run(); err != nil {
		// python3 venv not available, create one
		// check if system python3 is available
		if err := execCommand("python3", "--version"); err != nil {
			return "", errors.Wrap(err, "python3 not found")
		}

		// check if .venv exists
		if _, err := os.Stat(venvDir); err == nil {
			// .venv exists, remove first
			if err := execCommand("del", "/q", venvDir); err != nil {
				return "", errors.Wrap(err, "remove existed venv failed")
			}
		}

		// create python3 .venv
		// notice: --symlinks should be specified for windows
		// https://github.com/actions/virtual-environments/issues/2690
		if err := execCommand("python3", "-m", "venv", "--symlinks", venvDir); err != nil {
			// fix: failed to symlink on Windows
			log.Warn().Msg("failed to create python3 .venv by using --symlinks, try to use --copies")
			if err := execCommand("python3", "-m", "venv", "--copies", venvDir); err != nil {
				return "", errors.Wrap(err, "create python3 venv failed")
			}
		}

		// fix: python3 doesn't exist in .venv on Windows
		if _, err := os.Stat(python3); err != nil {
			log.Warn().Msg("python3 doesn't exist, try to link python")
			err := os.Link(filepath.Join(venvDir, "Scripts", "python.exe"), python3)
			if err != nil {
				return "", errors.Wrap(err, "python3 doesn't exist in .venv")
			}
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

func execCommand(cmdName string, args ...string) error {
	// "cmd /c" carries out the command specified by string and then stops
	// refer: https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/cmd
	cmdStr := fmt.Sprintf("%s %s", cmdName, strings.Join(args, " "))
	cmd := exec.Command("cmd", "/c", cmdStr)
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
