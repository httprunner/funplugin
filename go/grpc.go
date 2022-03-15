package hrpPlugin

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/httprunner/func-plugin/go/protoGen"
	"github.com/httprunner/func-plugin/shared"
	jsoniter "github.com/json-iterator/go"
)

// replace with third-party json library to improve performance
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// functionGRPCClient runs on the host side, it implements FuncCaller interface
type functionGRPCClient struct {
	client protoGen.DebugTalkClient
}

func (m *functionGRPCClient) GetNames() ([]string, error) {
	log.Info().Msg("function GetNames called on host side")
	resp, err := m.client.GetNames(context.Background(), &protoGen.Empty{})
	if err != nil {
		log.Error().Err(err).Msg("gRPC call GetNames() failed")
		return nil, err
	}
	return resp.Names, err
}

func (m *functionGRPCClient) Call(funcName string, funcArgs ...interface{}) (interface{}, error) {
	log.Info().Str("funcName", funcName).Interface("funcArgs", funcArgs).Msg("call function via gRPC")

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
		log.Error().Err(err).
			Str("funcName", funcName).Interface("funcArgs", funcArgs).
			Msg("gRPC Call() failed")
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
	log.Info().Interface("req", req).Msg("gRPC GetNames() called on plugin side")
	v, err := m.Impl.GetNames()
	if err != nil {
		log.Error().Err(err).Msg("gRPC GetNames() execution failed")
		return nil, err
	}
	return &protoGen.GetNamesResponse{Names: v}, err
}

func (m *functionGRPCServer) Call(ctx context.Context, req *protoGen.CallRequest) (*protoGen.CallResponse, error) {
	var funcArgs []interface{}
	if err := json.Unmarshal(req.Args, &funcArgs); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Call() funcArgs")
	}

	log.Info().Interface("req", req).Msg("gRPC Call() called on plugin side")

	v, err := m.Impl.Call(req.Name, funcArgs...)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("gRPC Call() execution failed")
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
