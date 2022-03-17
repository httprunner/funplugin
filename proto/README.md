# Updating the Protocol

If you update the protocol buffers file, you can regenerate the file using the following command from the project root directory. You do not need to run this if you're just using the plugin.

## For Go

### Install dependencies

ref: https://www.grpc.io/docs/languages/go/quickstart/

Install the protocol compiler plugins for Go using the following commands:

```bash
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Update your PATH so that the protoc compiler can find the plugins:

```bash
$ export PATH="$PATH:$(go env GOPATH)/bin"
```

### Generate gRPC code

```bash
$ protoc --go_out=. --go-grpc_out=. proto/debugtalk.proto
```

This will generate two go files in `go/protoGen` folder:

- debugtalk.pb.go
- debugtalk_grpc.pb.go

## For Python

### Install dependencies

ref: https://www.grpc.io/docs/languages/python/quickstart/

Install gRPC:

```bash
$ pip3 install grpcio
```

Install gRPC tools:

```bash
$ pip3 install grpcio-tools
```

Or you can just install all dependencies with `poetry`.

```bash
$ poetry install
```

### Generate gRPC code

```bash
$ python3 -m grpc_tools.protoc -I=proto --python_out=funppy/ --grpc_python_out=funppy/ proto/debugtalk.proto
```

This will generate two python files in `python` folder:

- debugtalk_pb2.py
- debugtalk_pb2_grpc.py

We need to modify `debugtalk_pb2_grpc.py` from `import debugtalk_pb2 as debugtalk__pb2` to `from funppy import debugtalk_pb2 as debugtalk__pb2`.

Or we can generate target files like this:

```bash
$ cp proto/debugtalk.proto funppy/debugtalk.proto
$ python3 -m grpc_tools.protoc -I=. --python_out=. --grpc_python_out=. funppy/debugtalk.proto
```
