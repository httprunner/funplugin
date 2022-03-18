
# go plugin

The golang official plugin is only supported on `Linux`, `FreeBSD`, and `macOS`. And this solution also has many drawbacks.

## create plugin functions

Firstly, you need to define your plugin functions. The functions can be very flexible, only the following restrictions should be complied with.

- plugin package name must be `main`.
- function names must be capitalized.
- function should return at most one value and one error.

Here is some plugin functions as example.

```go
package main

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
```

You can get more examples at [fungo/examples/debugtalk.go]

## build plugin

Then you can build your go plugin with `-buildmode=plugin` flag to binary file `xxx.so`. The file suffix of `.so` is by convention and should not be changed.

```bash
$ go build -buildmode=plugin -o=fungo/examples/xxx.so fungo/examples/debugtalk.go
```

## use plugin functions

Finally, you can use `Init` to initialize plugin via the `xxx.so` path, and you can call the plugin API to handle plugin functionality.

Notice: you should use the original function name.

[fungo/examples/debugtalk.go]: ../fungo/examples/debugtalk.go
