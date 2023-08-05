package fungo

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/httprunner/funplugin/fungo/protoGen"
	"github.com/httprunner/funplugin/shared"
	jsoniter "github.com/json-iterator/go"
)

// replace with third-party json library to improve performance
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// functionGRPCClient runs on the host side, it implements FuncCaller interface
type functionGRPCClient struct {
	client protoGen.DebugTalkClient
}

func (m *functionGRPCClient) GetNames() ([]string, error) {
	logger.Info("function GetNames called on host side")
	resp, err := m.client.GetNames(context.Background(), &protoGen.Empty{})
	if err != nil {
		logger.Error("gRPC call GetNames() failed", "error", err)
		return nil, err
	}
	return resp.Names, err
}

func (m *functionGRPCClient) Call(funcName string, funcArgs ...interface{}) (interface{}, error) {
	logger.Info("call function via gRPC", "funcName", funcName, "funcArgs", funcArgs)

	funcArgBytes, err := json.Marshal(funcArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal Call() funcArgs")
	}
	req := &protoGen.CallRequest{
		Name: funcName,
		Args: funcArgBytes,
	}

	response, err := m.client.Call(context.Background(), req)
	if err != nil {
		logger.Error("gRPC Call() failed",
			"funcName", funcName,
			"funcArgs", funcArgs,
			"error", err,
		)
		return nil, err
	}

	var resp interface{}
	err = json.Unmarshal(response.Value, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Call() response")
	}
	return resp, nil
}

// Here is the gRPC server that functionGRPCClient talks to.
type functionGRPCServer struct {
	protoGen.UnimplementedDebugTalkServer
	Impl shared.IFuncCaller
}

func (m *functionGRPCServer) GetNames(ctx context.Context, req *protoGen.Empty) (*protoGen.GetNamesResponse, error) {
	logger.Info("gRPC GetNames() called on plugin side", "req", req)
	v, err := m.Impl.GetNames()
	if err != nil {
		logger.Error("gRPC GetNames() execution failed", "error", err)
		return nil, err
	}
	return &protoGen.GetNamesResponse{Names: v}, err
}

func (m *functionGRPCServer) Call(ctx context.Context, req *protoGen.CallRequest) (*protoGen.CallResponse, error) {
	var funcArgs []interface{}
	if err := json.Unmarshal(req.Args, &funcArgs); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Call() funcArgs")
	}

	logger.Info("gRPC Call() called on plugin side", "req", req)

	v, err := m.Impl.Call(req.Name, funcArgs...)
	if err != nil {
		logger.Error("gRPC Call() execution failed", "req", req, "error", err)
		return nil, err
	}

	value, err := json.Marshal(v)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal Call() response")
	}
	return &protoGen.CallResponse{Value: value}, err
}

// GRPCPlugin implements hashicorp's plugin.GRPCPlugin.
type GRPCPlugin struct {
	plugin.Plugin
	Impl shared.IFuncCaller
}

func (p *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	protoGen.RegisterDebugTalkServer(s, &functionGRPCServer{Impl: p.Impl})
	return nil
}

func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &functionGRPCClient{client: protoGen.NewDebugTalkClient(c)}, nil
}
