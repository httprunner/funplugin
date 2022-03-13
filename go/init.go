package pluginInternal

import (
	"fmt"
	"os"
	"path/filepath"
)

type pluginFile string

const (
	goPluginFile          pluginFile = PluginName + ".so"  // built from go plugin
	hashicorpGoPluginFile pluginFile = PluginName + ".bin" // built from hashicorp go plugin
	hashicorpPyPluginFile pluginFile = PluginName + ".py"
)

func Init(path string, logOn bool) (IPlugin, error) {
	if path == "" {
		return nil, nil
	}
	var plugin IPlugin

	// priority: hashicorp plugin (debugtalk.bin > debugtalk.py) > go plugin (debugtalk.so)

	// locate hashicorp go/python plugin file
	hashicorpPluginFiles := []pluginFile{hashicorpGoPluginFile, hashicorpPyPluginFile}
	var pluginPath string
	var err error
	for _, hashicorpPluginFile := range hashicorpPluginFiles {
		pluginPath, err = locateFile(path, hashicorpPluginFile)
		if err == nil {
			break
		}
	}
	if err == nil {
		// found hashicorp go/python plugin file
		plugin = &HashicorpPlugin{
			logOn: logOn,
		}
		err = plugin.Init(pluginPath)
		return plugin, err
	}

	// locate go plugin file
	pluginPath, err = locateFile(path, goPluginFile)
	if err == nil {
		// found go plugin file
		plugin = &GoPlugin{}
		err = plugin.Init(pluginPath)
		return plugin, err
	}

	// plugin not found
	return nil, nil
}

// locateFile searches destFile upward recursively until current
// working directory or system root dir.
func locateFile(startPath string, destFile pluginFile) (string, error) {
	stat, err := os.Stat(startPath)
	if os.IsNotExist(err) {
		return "", err
	}

	var startDir string
	if stat.IsDir() {
		startDir = startPath
	} else {
		startDir = filepath.Dir(startPath)
	}
	startDir, _ = filepath.Abs(startDir)

	// convention over configuration
	pluginPath := filepath.Join(startDir, string(destFile))
	if _, err := os.Stat(pluginPath); err == nil {
		return pluginPath, nil
	}

	// current working directory
	cwd, _ := os.Getwd()
	if startDir == cwd {
		return "", fmt.Errorf("searched to CWD, plugin file not found")
	}

	// system root dir
	parentDir, _ := filepath.Abs(filepath.Dir(startDir))
	if parentDir == startDir {
		return "", fmt.Errorf("searched to system root dir, plugin file not found")
	}

	return locateFile(parentDir, destFile)
}
