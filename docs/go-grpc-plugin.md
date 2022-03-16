# Golang plugin over gRPC

It is recommended to use `golang plugin over gRPC` in most cases. This should be the most stable solution with higher performance.

## install SDK

Before you develop your golang plugin, you need to install an dependency as SDK.

```bash
$ go get github.com/httprunner/funplugin
```

## create plugin functions

Then you can write your plugin functions in golang. The functions can be very flexible, only the following restrictions should be complied with.

- package name should be `main`.
- function should return at most one value and one error.
- in `main()` function, `Register()` must be called to register plugin functions and `Serve()` must be called to start a plugin server process.

Here is some plugin functions as example.

```go
package main

import (
	"fmt"

	"github.com/httprunner/funplugin/fungo"
)

func SumTwoInt(a, b int) int {
	return a + b
}

func SumInts(args ...int) int {
	var sum int
	for _, arg := range args {
		sum += arg
	}
	return sum
}

func Sum(args ...interface{}) (interface{}, error) {
	var sum float64
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			sum += float64(v)
		case float64:
			sum += v
		default:
			return nil, fmt.Errorf("unexpected type: %T", arg)
		}
	}
	return sum, nil
}

func main() {
	fungo.Register("sum_ints", SumInts)
	fungo.Register("sum_two_int", SumTwoInt)
	fungo.Register("sum", Sum)
	fungo.Serve()
}
```

You can get more examples at [fungo/examples/].

## build plugin

Once the plugin functions are ready, you can build them into the binary file `debugtalk.bin`. The name of `debugtalk.bin` is by convention and should not be changed.

```bash
$ go build -o fungo/examples/debugtalk.bin fungo/examples/hashicorp.go fungo/examples/debugtalk.go
```

## use plugin functions

Finally, you can use `Init` to initialize plugin via the `debugtalk.bin` path, and you can call the plugin API to handle plugin functionality.


[fungo/examples/]: ../fungo/examples/
