package hrpPlugin

func Init(path string, logOn bool) (IPlugin, error) {
	if path == "" {
		return nil, nil
	}
	var plugin IPlugin

	// priority: hashicorp plugin (debugtalk.bin > debugtalk.py) > go plugin (debugtalk.so)

	// locate hashicorp go/python plugin file
	hashicorpPluginFiles := []string{hashicorpGoPluginFile, hashicorpPyPluginFile}
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
		plugin = &hashicorpPlugin{
			logOn: logOn,
		}
		err = plugin.Init(pluginPath)
		return plugin, err
	}

	// locate go plugin file
	pluginPath, err = locateFile(path, goPluginFile)
	if err == nil {
		// found go plugin file
		plugin = &goPlugin{}
		err = plugin.Init(pluginPath)
		return plugin, err
	}

	// plugin not found
	return nil, nil
}
