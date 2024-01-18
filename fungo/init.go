package fungo

import (
	"io"
	"os"
	"path/filepath"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

const Version = "v0.5.4"

var (
	logger = Logger
)

var Logger = hclog.New(&hclog.LoggerOptions{
	Name:        "fungo",
	Output:      hclog.DefaultOutput,
	DisableTime: true,
	Level:       hclog.Debug,
	Color:       hclog.AutoColor,
})

var file *os.File

func InitLogger(logLevel hclog.Level, logFile string, disableTime bool) hclog.Logger {
	output := hclog.DefaultOutput
	if logFile != "" {
		err := os.MkdirAll(filepath.Dir(logFile), os.ModePerm)
		if err != nil {
			logger.Error("create log file directory failed",
				"error", err, "logFile", logFile)
			os.Exit(1)
		}

		file, err = os.OpenFile(logFile, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			logger.Error("open log file failed", "error", err)
			os.Exit(1)
		}
		output = io.MultiWriter(hclog.DefaultOutput, file)
	}

	logger = hclog.New(&hclog.LoggerOptions{
		Name:        "fungo",
		Output:      output,
		DisableTime: disableTime,
		Level:       logLevel,
		Color:       hclog.AutoColor,
	})
	logger.Info("set plugin log level",
		"level", logLevel.String(), "logFile", logFile)
	return logger
}

func CloseLogFile() error {
	if file != nil {
		logger.Info("close log file")
		return file.Close()
	}
	return nil
}

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
