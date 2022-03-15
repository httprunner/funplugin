package funcPlugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	funcGoPlugin "github.com/httprunner/func-plugin/go"
	"github.com/httprunner/func-plugin/shared"
)

// hashicorpPlugin implements hashicorp/go-plugin
type hashicorpPlugin struct {
	client          *plugin.Client
	pluginType      string
	funcCaller      shared.IFuncCaller
	logOn           bool            // turn on plugin log
	cachedFunctions map[string]bool // cache loaded functions to improve performance
}

func (p *hashicorpPlugin) Init(path string) error {
	pluginType := os.Getenv(shared.PluginTypeEnvName)
	if pluginType == "rpc" {
		p.pluginType = pluginType
	} else {
		p.pluginType = "grpc" // default
	}
	pluginName := shared.PluginName + "_" + p.pluginType

	// logger
	loggerOptions := &hclog.LoggerOptions{
		Name:   pluginName,
		Output: os.Stdout,
	}
	if p.logOn {
		loggerOptions.Level = hclog.Debug
	} else {
		loggerOptions.Level = hclog.Info
	}

	// cmd
	var cmd *exec.Cmd
	if filepath.Base(path) == shared.HashicorpPyPluginFile {
		// hashicorp python plugin
		cmd = exec.Command("python3", path)
	} else {
		// go plugin
		cmd = exec.Command(path)
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", shared.PluginTypeEnvName, p.pluginType))

	// launch the plugin process
	p.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			shared.PluginName + "_rpc":  &funcGoPlugin.RPCPlugin{},
			shared.PluginName + "_grpc": &funcGoPlugin.GRPCPlugin{},
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
		return errors.Wrap(err, fmt.Sprintf("connect %s plugin failed", p.pluginType))
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("request %s plugin failed", p.pluginType))
	}

	// We should have a Function now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	p.funcCaller = raw.(shared.IFuncCaller)

	p.cachedFunctions = make(map[string]bool)
	log.Info().Str("path", path).Msg("load hashicorp go plugin success")
	return nil
}

func (p *hashicorpPlugin) Type() string {
	return fmt.Sprintf("hashicorp-%s", p.pluginType)
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
