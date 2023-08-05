package funplugin

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"

	"github.com/httprunner/funplugin/fungo"
	"github.com/httprunner/funplugin/shared"
)

type rpcType string

const (
	rpcTypeRPC  rpcType = "rpc"  // go net/rpc
	rpcTypeGRPC rpcType = "grpc" // default
)

func (t rpcType) String() string {
	return string(t)
}

// hashicorpPlugin implements hashicorp/go-plugin
type hashicorpPlugin struct {
	client          *plugin.Client
	rpcType         rpcType
	funcCaller      shared.IFuncCaller
	cachedFunctions sync.Map // cache loaded functions to improve performance, key is function name, value is bool
	path            string   // plugin file path
	option          *pluginOption
}

func newHashicorpPlugin(path string, option *pluginOption) (*hashicorpPlugin, error) {
	p := &hashicorpPlugin{
		path:   path,
		option: option,
	}

	// plugin type, grpc or rpc
	p.rpcType = rpcType(os.Getenv(shared.PluginTypeEnvName))
	if p.rpcType != rpcTypeRPC {
		p.rpcType = rpcTypeGRPC // default
	}

	// logger
	logger = logger.Named(fmt.Sprintf("%v-%v", p.rpcType, p.option.langType))
	if p.option.debugLogger {
		logger.Info("set plugin log level to DEBUG")
		logger.SetLevel(hclog.Debug)
	} else {
		logger.Info("set plugin log level to INFO")
		logger.SetLevel(hclog.Info)
	}

	// cmd
	var cmd *exec.Cmd
	if p.option.langType == langTypePython {
		// hashicorp python plugin
		cmd = exec.Command(p.option.python3, path)
	} else {
		// hashicorp go plugin
		cmd = exec.Command(path)
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", shared.PluginTypeEnvName, p.rpcType))

	// launch the plugin process
	logger.Info("launch the plugin process")
	p.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			rpcTypeRPC.String():  &fungo.RPCPlugin{},
			rpcTypeGRPC.String(): &fungo.GRPCPlugin{},
		},
		Cmd:    cmd,
		Logger: logger,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC,
			plugin.ProtocolGRPC,
		},
	})

	// Connect via RPC/gRPC
	rpcClient, err := p.client.Client()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("connect %s plugin failed", p.rpcType))
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(p.rpcType.String())
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("request %s plugin failed", p.rpcType))
	}

	// We should have a Function now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	p.funcCaller = raw.(shared.IFuncCaller)

	p.cachedFunctions = sync.Map{}
	logger.Info("load hashicorp go plugin success", "path", path)

	return p, nil
}

func (p *hashicorpPlugin) Type() string {
	return fmt.Sprintf("hashicorp-%s-%v", p.rpcType, p.option.langType)
}

func (p *hashicorpPlugin) Path() string {
	return p.path
}

func (p *hashicorpPlugin) Has(funcName string) bool {
	logger.Debug("check if plugin has function", "funcName", funcName)
	flag, ok := p.cachedFunctions.Load(funcName)
	if ok {
		return flag.(bool)
	}

	funcNames, err := p.funcCaller.GetNames()
	if err != nil {
		return false
	}

	for _, name := range funcNames {
		if name == funcName {
			p.cachedFunctions.Store(funcName, true) // cache as exists
			return true
		}
	}

	p.cachedFunctions.Store(funcName, false) // cache as not exists
	return false
}

func (p *hashicorpPlugin) Call(funcName string, args ...interface{}) (interface{}, error) {
	return p.funcCaller.Call(funcName, args...)
}

func (p *hashicorpPlugin) Quit() error {
	// kill hashicorp plugin process
	logger.Info("quit hashicorp plugin process")
	p.client.Kill()
	return nil
}
