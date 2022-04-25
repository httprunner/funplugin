package shared

import (
	"github.com/hashicorp/go-plugin"
)

const Version = "v0.4.4"

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
