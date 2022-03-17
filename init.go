package funplugin

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type IPlugin interface {
	Type() string                                                   // get plugin type
	Has(funcName string) bool                                       // check if plugin has function
	Call(funcName string, args ...interface{}) (interface{}, error) // call function
	Quit() error                                                    // quit plugin
}

// Init initializes plugin with plugin path
func Init(path string, logOn bool) (plugin IPlugin, err error) {
	// priority: hashicorp plugin (debugtalk.bin > debugtalk.py) > go plugin (debugtalk.so)
	ext := filepath.Ext(path)
	switch ext {
	case ".bin":
		// found hashicorp go plugin file
		return newHashicorpPlugin(path, logOn, "")
	case ".py":
		// found hashicorp python plugin file
		python3, err := preparePython3Venv(path)
		if err != nil {
			log.Error().Err(err).Msg("prepare python venv failed")
			return nil, err
		}
		return newHashicorpPlugin(path, logOn, python3)
	case ".so":
		// found go plugin file
		return newGoPlugin(path)
	default:
		log.Error().Err(err).Msgf("invalid plugin path: %s", path)
		return nil, fmt.Errorf("unsupported plugin type: %s", ext)
	}
}

// preparePython3Venv prepares python3 venv for hashicorp python plugin
func preparePython3Venv(path string) (python3 string, err error) {
	projectDir := filepath.Dir(path)
	if err := execCommand(exec.Command("python3", "--version"), projectDir); err != nil {
		return "", errors.Wrap(err, "python3 not found")
	}
	if err := execCommand(exec.Command("python3", "-m", "venv", ".venv"), projectDir); err != nil {
		return "", errors.Wrap(err, "create python3 venv failed")
	}
	python3, _ = filepath.Abs(filepath.Join(projectDir, ".venv", "bin", "python3"))
	pip3InstallCmd := exec.Command(python3, "-m",
		"pip", "--disable-pip-version-check", "install", "funppy")
	if err := execCommand(pip3InstallCmd, projectDir); err != nil {
		return "", errors.Wrap(err, "install funppy failed")
	}
	return python3, nil
}

func execCommand(cmd *exec.Cmd, cwd string) error {
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
