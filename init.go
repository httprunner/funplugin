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

type langType string

const (
	langTypeGo     langType = "go"
	langTypePython langType = "py"
	langTypeJava   langType = "java"
)

type pluginOption struct {
	logOn    bool
	langType langType // go or py
	python3  string   // python3 path with funppy dependency
}

type Option func(*pluginOption)

func WithLogOn(logOn bool) Option {
	return func(o *pluginOption) {
		o.logOn = logOn
	}
}

func WithPython3(python3 string) Option {
	return func(o *pluginOption) {
		o.python3 = python3
	}
}

// Init initializes plugin with plugin path
func Init(path string, options ...Option) (plugin IPlugin, err error) {
	option := &pluginOption{}
	for _, o := range options {
		o(option)
	}

	// priority: hashicorp plugin > go plugin
	ext := filepath.Ext(path)
	switch ext {
	case ".bin":
		// found hashicorp go plugin file
		option.langType = langTypeGo
		return newHashicorpPlugin(path, option)
	case ".py":
		// found hashicorp python plugin file
		if option.python3 == "" {
			python3, err := shared.PreparePython3Venv(path)
			if err != nil {
				log.Error().Err(err).Msg("prepare python venv failed")
				return nil, err
			}
			option.python3 = python3
		}
		option.langType = langTypePython
		return newHashicorpPlugin(path, option)
	case ".so":
		// found go plugin file
		return newGoPlugin(path)
	default:
		log.Error().Err(err).Msgf("invalid plugin path: %s", path)
		return nil, fmt.Errorf("unsupported plugin type: %s", ext)
	}
}
