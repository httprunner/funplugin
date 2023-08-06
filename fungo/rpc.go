package fungo

import (
	"encoding/gob"
	"net/rpc"

	"github.com/hashicorp/go-plugin"

	"github.com/httprunner/funplugin/shared"
)

func init() {
	gob.Register(new(funcData))
}

// funcData is used to transfer between plugin and host via RPC.
type funcData struct {
	Name string        // function name
	Args []interface{} // function arguments
}

// functionRPCClient runs on the host side, it implements FuncCaller interface
type functionRPCClient struct {
	client *rpc.Client
}

func (g *functionRPCClient) GetNames() ([]string, error) {
	logger.Debug("rpc_client GetNames() start")
	var resp []string
	err := g.client.Call("Plugin.GetNames", new(interface{}), &resp)
	if err != nil {
		logger.Error("rpc_client GetNames() failed", "error", err)
		return nil, err
	}
	logger.Debug("rpc_client GetNames() success")
	return resp, nil
}

// host -> plugin
func (g *functionRPCClient) Call(funcName string, funcArgs ...interface{}) (interface{}, error) {
	logger.Info("rpc_client Call() start", "funcName", funcName, "funcArgs", funcArgs)
	f := funcData{
		Name: funcName,
		Args: funcArgs,
	}

	var args interface{} = f
	var resp interface{}
	err := g.client.Call("Plugin.Call", &args, &resp)
	if err != nil {
		logger.Error("rpc_client Call() failed",
			"funcName", funcName,
			"funcArgs", funcArgs,
			"error", err,
		)
		return nil, err
	}
	logger.Info("rpc_client Call() success", "result", resp)
	return resp, nil
}

// functionRPCServer runs on the plugin side, executing the user custom function.
type functionRPCServer struct {
	Impl shared.IFuncCaller
}

// plugin execution
func (s *functionRPCServer) GetNames(args interface{}, resp *[]string) error {
	logger.Debug("rpc_server GetNames() start")
	var err error
	*resp, err = s.Impl.GetNames()
	if err != nil {
		logger.Error("rpc_server GetNames() failed", "error", err)
		return err
	}
	logger.Debug("rpc_server GetNames() success")
	return nil
}

// plugin execution
func (s *functionRPCServer) Call(args interface{}, resp *interface{}) error {
	logger.Debug("rpc_server Call() start")
	f := args.(*funcData)
	var err error
	*resp, err = s.Impl.Call(f.Name, f.Args...)
	if err != nil {
		logger.Error("rpc_server Call() failed", "args", args, "error", err)
		return err
	}
	logger.Debug("rpc_server Call() success")
	return nil
}

// RPCPlugin implements hashicorp's plugin.Plugin.
type RPCPlugin struct {
	Impl shared.IFuncCaller
}

func (p *RPCPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &functionRPCServer{Impl: p.Impl}, nil
}

func (p *RPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &functionRPCClient{client: c}, nil
}
