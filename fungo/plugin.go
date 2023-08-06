package fungo

import (
	"fmt"
	"os"
	"reflect"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// functionsMap stores plugin functions
type functionsMap map[string]reflect.Value

// functionPlugin implements the FuncCaller interface
type functionPlugin struct {
	logger    hclog.Logger
	functions functionsMap
}

func (p *functionPlugin) GetNames() ([]string, error) {
	var names []string
	for name := range p.functions {
		names = append(names, name)
	}
	p.logger.Debug("get registered plugin functions", "names", names)
	return names, nil
}

func (p *functionPlugin) Call(funcName string, args ...interface{}) (interface{}, error) {
	// notice: this is the actual place where plugin function is called
	p.logger.Debug("plugin function execution", "funcName", funcName, "args", args)

	fn, ok := p.functions[funcName]
	if !ok {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	return CallFunc(fn, args...)
}

var functions = make(functionsMap)

// Register registers a plugin function.
// Every plugin function must be registered before Serve() is called.
func Register(funcName string, fn interface{}) {
	if _, ok := functions[funcName]; ok {
		return
	}
	logger.Info("register plugin function", "funcName", funcName)
	functions[funcName] = reflect.ValueOf(fn)
	// automatic registration with common name
	functions[ConvertCommonName(funcName)] = functions[funcName]
}

// serveRPC starts a plugin server process in RPC mode.
func serveRPC() {
	rpcPluginName := "rpc"
	logger.Info("start plugin server in RPC mode")
	funcPlugin := &functionPlugin{
		logger:    logger.Named("func_exec"),
		functions: functions,
	}
	var pluginMap = map[string]plugin.Plugin{
		rpcPluginName: &RPCPlugin{Impl: funcPlugin},
	}
	// start RPC server
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         pluginMap,
	})
}

// serveGRPC starts a plugin server process in gRPC mode.
func serveGRPC() {
	grpcPluginName := "grpc"
	logger.Info("start plugin server in gRPC mode")
	funcPlugin := &functionPlugin{
		logger:    logger.Named("func_exec"),
		functions: functions,
	}
	var pluginMap = map[string]plugin.Plugin{
		grpcPluginName: &GRPCPlugin{Impl: funcPlugin},
	}
	// start gRPC server
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}

// default to run plugin in gRPC mode
func Serve() {
	if os.Getenv(PluginTypeEnvName) == "rpc" {
		serveRPC()
	} else {
		// default
		serveGRPC()
	}
}
