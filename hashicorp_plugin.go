package funplugin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

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
	cachedFunctions map[string]bool // cache loaded functions to improve performance
	option          *pluginOption
}

func newHashicorpPlugin(path string, option *pluginOption) (*hashicorpPlugin, error) {
	p := &hashicorpPlugin{
		option: option,
	}

	// plugin type, grpc or rpc
	p.rpcType = rpcType(os.Getenv(shared.PluginTypeEnvName))
	if p.rpcType != rpcTypeRPC {
		p.rpcType = rpcTypeGRPC // default
	}

	// logger
	loggerOptions := &hclog.LoggerOptions{
		Name:   fmt.Sprintf("%v-%v", p.rpcType, p.option.langType),
		Output: os.Stdout,
	}
	if p.option.logOn {
		// turn on plugin log
		loggerOptions.Level = hclog.Debug
	} else {
		loggerOptions.Level = hclog.Info
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
	p.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			rpcTypeRPC.String():  &fungo.RPCPlugin{},
			rpcTypeGRPC.String(): &fungo.GRPCPlugin{},
		},
		Cmd:    cmd,
		Logger: hclog.New(loggerOptions),
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

	p.cachedFunctions = make(map[string]bool)
	log.Info().Str("path", path).Msg("load hashicorp go plugin success")

	return p, nil
}

func (p *hashicorpPlugin) Type() string {
	return fmt.Sprintf("hashicorp-%s-%v", p.rpcType, p.option.langType)
}

func (p *hashicorpPlugin) Has(funcName string) bool {
	flag, ok := p.cachedFunctions[funcName]
	if ok {
		return flag
	}

	funcNames, err := p.funcCaller.GetNames()
	if err != nil {
		return false
	}

	for _, name := range funcNames {
		if name == funcName {
			p.cachedFunctions[funcName] = true // cache as exists
			return true
		}
	}

	p.cachedFunctions[funcName] = false // cache as not exists
	return false
}

func (p *hashicorpPlugin) Call(funcName string, args ...interface{}) (interface{}, error) {
	return p.funcCaller.Call(funcName, args...)
}

func (p *hashicorpPlugin) Quit() error {
	// kill hashicorp plugin process
	log.Info().Msg("quit hashicorp plugin process")
	p.client.Kill()
	return nil
}
