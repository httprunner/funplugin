package funplugin

import (
	"github.com/httprunner/funplugin/shared"
)

type IPlugin interface {
	Init(path string) error                                         // init plugin
	Type() string                                                   // get plugin type
	Has(funcName string) bool                                       // check if plugin has function
	Call(funcName string, args ...interface{}) (interface{}, error) // call function
	Quit() error                                                    // quit plugin
}

func Init(path string, logOn bool) (plugin IPlugin, err error) {
	if path == "" {
		return nil, nil
	}

	// priority: hashicorp plugin (debugtalk.bin > debugtalk.py) > go plugin (debugtalk.so)

	// locate hashicorp go/python plugin file
	hashicorpPluginFiles := []string{shared.HashicorpGoPluginFile, shared.HashicorpPyPluginFile}
	var pluginPath string
	for _, hashicorpPluginFile := range hashicorpPluginFiles {
		pluginPath, err = shared.LocateFile(path, hashicorpPluginFile)
		if err == nil {
			break
		}
	}
	if err == nil {
		// found hashicorp go/python plugin file
		plugin = &hashicorpPlugin{
			logOn: logOn,
		}
		err = plugin.Init(pluginPath)
		return plugin, err
	}

	// locate go plugin file
	pluginPath, err = shared.LocateFile(path, shared.GoPluginFile)
	if err == nil {
		// found go plugin file
		plugin = &goPlugin{}
		err = plugin.Init(pluginPath)
		return plugin, err
	}

	// plugin not found
	return nil, nil
}
