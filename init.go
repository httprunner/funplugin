package funplugin

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/httprunner/funplugin/fungo"
	"github.com/httprunner/funplugin/myexec"
)

var (
	logger = fungo.Logger
)

type IPlugin interface {
	Type() string                                                   // get plugin type
	Path() string                                                   // get plugin file path
	Has(funcName string) bool                                       // check if plugin has function
	Call(funcName string, args ...interface{}) (interface{}, error) // call function
	Quit() error                                                    // quit plugin
	StartHeartbeat()                                                // heartbeat to keep the plugin alive
}

type langType string

const (
	langTypeGo     langType = "go"
	langTypePython langType = "py"
	langTypeJava   langType = "java"
)

type pluginOption struct {
	debugLogger    bool     // whether set log level to DEBUG
	logFile        string   // specify log file path
	disableLogTime bool     // whether disable log time
	langType       langType // go or py
	python3        string   // python3 path with funppy dependency
}

type Option func(*pluginOption)

func WithDebugLogger(debug bool) Option {
	return func(o *pluginOption) {
		o.debugLogger = debug
	}
}

func WithLogFile(logFile string) Option {
	return func(o *pluginOption) {
		o.logFile = logFile
	}
}

func WithDisableTime(disable bool) Option {
	return func(o *pluginOption) {
		o.disableLogTime = disable
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

	// init logger
	logLevel := hclog.Info
	if option.debugLogger {
		logLevel = hclog.Debug
	}
	logger = fungo.InitLogger(
		logLevel, option.logFile, option.disableLogTime)

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
			// create python3 venv with funppy if python3 not specified
			option.python3, err = myexec.EnsurePython3Venv("", "funppy")
			if err != nil {
				logger.Error("prepare python3 funppy venv failed", "error", err)
				return nil, errors.Wrap(err,
					"miss python3, create python3 funppy venv failed")
			}
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
