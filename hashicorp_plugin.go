package funplugin

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"

	"github.com/httprunner/funplugin/fungo"
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
	funcCaller      fungo.IFuncCaller
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
	p.rpcType = rpcType(os.Getenv(fungo.PluginTypeEnvName))
	if p.rpcType != rpcTypeRPC {
		p.rpcType = rpcTypeGRPC // default
	}
	// logger
	logger = logger.ResetNamed(fmt.Sprintf("hc-%v-%v", p.rpcType, p.option.langType))

	// 失败则继续尝试，连续三次失败则返回错误
	err := p.startPlugin()
	if err == nil {
		return p, err
	}
	logger.Info("load hashicorp go plugin success", "path", path)

	return nil, err
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

func (p *hashicorpPlugin) StartHeartbeat() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	var err error

	for range ticker.C {
		// Check the client connection status
		logger.Info("heartbreak......")
		if p.client.Exited() {
			logger.Error(fmt.Sprintf("plugin exited, restarting..."))
			err = p.startPlugin()
			if err != nil {
				break
			}
		}
	}
}

func (p *hashicorpPlugin) startPlugin() error {
	var cmd *exec.Cmd
	if p.option.langType == langTypePython {
		// hashicorp python plugin
		cmd = exec.Command(p.option.python3, p.path)
		// hashicorp python plugin only supports gRPC
		p.rpcType = rpcTypeGRPC
	} else {
		// hashicorp go plugin
		cmd = exec.Command(p.path)
		// hashicorp go plugin supports grpc and rpc
		p.rpcType = rpcType(os.Getenv(fungo.PluginTypeEnvName))
		if p.rpcType != rpcTypeRPC {
			p.rpcType = rpcTypeGRPC // default
		}
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", fungo.PluginTypeEnvName, p.rpcType))

	var err error
	maxRetryCount := 3
	for i := 0; i < maxRetryCount; i++ {
		err = p.tryStartPlugin(cmd, logger)
		if err == nil {
			return nil
		}
		time.Sleep(time.Second * time.Duration(i*i)) // sleep temporarily before next try
	}
	logger.Error("failed to start plugin after max retries")
	return errors.Wrap(err, "failed to start plugin after max retries")
}

func (p *hashicorpPlugin) tryStartPlugin(cmd *exec.Cmd, logger hclog.Logger) error {
	// launch the plugin process
	p.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: fungo.HandshakeConfig,
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
		return errors.Wrap(err, fmt.Sprintf("connect %s plugin failed", p.rpcType))
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(p.rpcType.String())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("request %s plugin failed", p.rpcType))
	}

	// We should have a Function now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	p.funcCaller = raw.(fungo.IFuncCaller)

	p.cachedFunctions = sync.Map{}

	return nil
}

func (p *hashicorpPlugin) Quit() error {
	// kill hashicorp plugin process
	logger.Info("quit hashicorp plugin process")
	p.client.Kill()
	return fungo.CloseLogFile()
}
