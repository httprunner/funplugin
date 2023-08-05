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
	var resp []string
	err := g.client.Call("Plugin.GetNames", new(interface{}), &resp)
	if err != nil {
		logger.Error("rpc call GetNames() failed", "error", err)
		return nil, err
	}
	return resp, nil
}

// host -> plugin
func (g *functionRPCClient) Call(funcName string, funcArgs ...interface{}) (interface{}, error) {
	logger.Info("call function via RPC", "funcName", funcName, "funcArgs", funcArgs)
	f := funcData{
		Name: funcName,
		Args: funcArgs,
	}

	var args interface{} = f
	var resp interface{}
	err := g.client.Call("Plugin.Call", &args, &resp)
	if err != nil {
		logger.Error("rpc Call() failed",
			"funcName", funcName,
			"funcArgs", funcArgs,
			"error", err,
		)
		return nil, err
	}
	return resp, nil
}

// functionRPCServer runs on the plugin side, executing the user custom function.
type functionRPCServer struct {
	Impl shared.IFuncCaller
}

// plugin execution
func (s *functionRPCServer) GetNames(args interface{}, resp *[]string) error {
	logger.Info("rpc GetNames() called on plugin side", "args", args)
	var err error
	*resp, err = s.Impl.GetNames()
	if err != nil {
		logger.Error("rpc GetNames() execution failed", "error", err)
		return err
	}
	return nil
}

// plugin execution
func (s *functionRPCServer) Call(args interface{}, resp *interface{}) error {
	logger.Info("rpc Call() called on plugin side", "args", args)
	f := args.(*funcData)
	var err error
	*resp, err = s.Impl.Call(f.Name, f.Args...)
	if err != nil {
		logger.Error("rpc Call() execution failed", "args", args, "error", err)
		return err
	}
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
