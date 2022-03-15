package hrpPlugin

import (
	"os"
	"strings"

	"github.com/hashicorp/go-plugin"
)

const (
	pluginName            = "debugtalk"
	rpcPluginName         = pluginName + "_rpc"
	grpcPluginName        = pluginName + "_grpc"
	goPluginFile          = pluginName + ".so"  // built from go plugin
	hashicorpGoPluginFile = pluginName + ".bin" // built from hashicorp go plugin
	hashicorpPyPluginFile = pluginName + ".py"
)

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "HttpRunnerPlus",
	MagicCookieValue: pluginName,
}

const hrpPluginTypeEnvName = "HRP_PLUGIN_TYPE"

var hrpPluginType string

func init() {
	hrpPluginType = strings.ToLower(os.Getenv(hrpPluginTypeEnvName))
	if hrpPluginType == "" {
		hrpPluginType = "grpc" // default
	}
}

func isRPCPluginType() bool {
	return hrpPluginType == "rpc"
}
