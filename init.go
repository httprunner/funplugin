package funplugin

import (
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/funplugin/shared"
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
		python3, err := shared.PreparePython3Venv(path)
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
