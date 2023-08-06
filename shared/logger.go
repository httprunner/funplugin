package shared

import "github.com/hashicorp/go-hclog"

var Logger = hclog.New(&hclog.LoggerOptions{
	Name:   "funplugin",
	Output: hclog.DefaultOutput,
	Level:  hclog.Debug,
	Color:  hclog.AutoColor,
})
