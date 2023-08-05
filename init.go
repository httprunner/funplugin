package funplugin

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/httprunner/funplugin/shared"
)

var (
	logger = shared.Logger
)

type IPlugin interface {
	Type() string                                                   // get plugin type
	Path() string                                                   // get plugin file path
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
	debugLogger bool
	langType    langType // go or py
	python3     string   // python3 path with funppy dependency
}

type Option func(*pluginOption)

func WithDebugLogger(debug bool) Option {
	return func(o *pluginOption) {
		o.debugLogger = debug
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
	logger.Info("init plugin", "path", path)

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
			return nil, errors.New("python3 not specified")
		}
		option.langType = langTypePython
		return newHashicorpPlugin(path, option)
	case ".so":
		// found go plugin file
		return newGoPlugin(path)
	default:
		logger.Error("invalid plugin path", "path", path, "error", err)
		return nil, fmt.Errorf("unsupported plugin type: %s", ext)
	}
}
