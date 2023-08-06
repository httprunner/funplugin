package fungo

import (
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

const Version = "v0.5.2"

var (
	logger = Logger
)

var Logger = hclog.New(&hclog.LoggerOptions{
	Name:        "fungo",
	Output:      hclog.DefaultOutput,
	DisableTime: false,
	Level:       hclog.Debug,
	Color:       hclog.AutoColor,
})

// PluginTypeEnvName is used to specify hashicorp go plugin type, rpc/grpc
const PluginTypeEnvName = "HRP_PLUGIN_TYPE"

// HandshakeConfig is used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "HttpRunnerPlus",
	MagicCookieValue: "debugtalk",
}

// IFuncCaller is the interface that we're exposing as a plugin.
type IFuncCaller interface {
	GetNames() ([]string, error)                                    // get all plugin function names list
	Call(funcName string, args ...interface{}) (interface{}, error) // call plugin function
}
