# Golang plugin over net/rpc

Using `golang plugin over net/rpc` is basically the same as [golang plugin over gRPC].

The only difference is that if you want to run the plugin in `net/rpc` mode, you need to set an environment variable `HRP_PLUGIN_TYPE=rpc`.

Set environment variable in shell:

```bash
$ export HRP_PLUGIN_TYPE=rpc
```

Or in your golang code:

```go
os.Setenv("HRP_PLUGIN_TYPE", "rpc")
```


[golang plugin over gRPC]: go-grpc-plugin.md
[examples/plugin/]: ../examples/plugin/
[examples/plugin/debugtalk.go]: ../examples/plugin/debugtalk.go
