# FunPlugin

[![Go Reference](https://pkg.go.dev/badge/github.com/httprunner/funplugin.svg)](https://pkg.go.dev/github.com/httprunner/funplugin)
[![Github Actions](https://github.com/httprunner/funplugin/actions/workflows/unittest.yml/badge.svg)](https://github.com/httprunner/funplugin/actions)
[![codecov](https://codecov.io/gh/httprunner/funplugin/branch/main/graph/badge.svg?token=DW3K2R1PNC)](https://codecov.io/gh/httprunner/funplugin)
[![Go Report Card](https://goreportcard.com/badge/github.com/httprunner/funplugin)](https://goreportcard.com/report/github.com/httprunner/funplugin)

## What is FunPlugin?

`FunPlugin` is short for function plugin, and I hope you can have fun using this plugin too.

This plugin project comes from the requirements of [HttpRunner+], because we need to implement some dynamic calculations or custom logic processing in YAML/JSON test cases. If you have used [HttpRunner] before, you must be impressed by `debugtalk.py`, which allows us to customize functions very easily and reference them in plain text test cases.

As the HttpRunner project evolves, we expect users to have more choices besides Python when writing custom functions, such as Golang, Java, Node, C++, C#, etc.

`FunPlugin` achieves this goal well and grows into an independent project that not only serves HttpRunner+, but can also be easily integrated into other golang projects.

## How to use FunPlugin?

`FunPlugin` is mainly based on [hashicorp plugin], which is a golang plugin system over RPC. It supports serving plugins via `gRPC`, which means plugins can be written in any language. You can find the official programming languages supported by grpc [here][grpc-lang].

Integrating `FunPlugin` is very simple. You only need to focus on two parts.

- client side: integrate FunPlugin into your golang project, call plugin functions at will.
- plugin side: write plugin functions in your favorite language and build them to plugin binary

### client call

FunPlugin has a very concise golang API that can be easily integrated into golang projects.

1, use `Init` to initialize plugin via plugin path.

```go
func Init(path string, options ...Option) (plugin IPlugin, err error)
```

- path: built plugin file path
- options: specify extra plugin options
  - WithLogOn(logOn bool): whether to print logs in plugin functions
  - WithPython3(python3 string): specify custom python3 path

2, call plugin API to deal with plugin functions.

If the specified plugin path is valid, you will get a plugin instance conforming to the `IPlugin` interface.

```go
type IPlugin interface {
	Type() string
	Has(funcName string) bool
	Call(funcName string, args ...interface{}) (interface{}, error)
	Quit() error
}
```

- Type: returns plugin type
- Has: check if plugin has a function
- Call: call function with function name and arguments
- Quit: quit plugin

You can reference [hashicorp_plugin_test.go] and [go_plugin_test.go] as examples.

### plugin server

In `RPC` architecture, plugins can be considered as servers. You can write plugin functions in your favorite language and then build them to a binary file. When the client `Init` the plugin file path, it starts the plugin as a server and they can then communicates via RPC.

Currently, `FunPlugin` supports 3 different plugins via RPC. You can check their documentation for more details.

- [x] [Golang plugin over gRPC][go-grpc-plugin], built as `xxx.bin` (recommended)
- [x] [Golang plugin over net/rpc][go-rpc-plugin], built as `xxx.bin`
- [x] [Python plugin over gRPC][python-grpc-plugin], no need to build, just name it with `xxx.py`

You are welcome to contribute more plugins in other languages.

- [ ] Java plugin over gRPC
- [ ] Node plugin over gRPC
- [ ] C++ plugin over gRPC
- [ ] C# plugin over gRPC
- [ ] [etc.][grpc-lang]

Finally, `FunPlugin` also supports writing plugin function with the official [go plugin]. However, this solution has a number of limitations. You can check this [document][go-plugin] for more details.


[HttpRunner+]: https://github.com/httprunner/hrp
[HttpRunner]: https://github.com/httprunner/httprunner
[hashicorp plugin]: https://github.com/hashicorp/go-plugin
[grpc-lang]: https://www.grpc.io/docs/languages/
[go plugin]: https://pkg.go.dev/plugin
[examples/plugin/]: ../examples/plugin/
[examples/plugin/debugtalk.go]: ../examples/plugin/debugtalk.go
[hashicorp_plugin_test.go]: hashicorp_plugin_test.go
[go_plugin_test.go]: go_plugin_test.go
[go-grpc-plugin]: docs/go-grpc-plugin.md
[go-rpc-plugin]: docs/go-rpc-plugin.md
[python-grpc-plugin]: docs/python-grpc-plugin.md
[go-plugin]: docs/go-plugin.md
