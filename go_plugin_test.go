//go:build linux || freebsd || darwin
// +build linux freebsd darwin

// go plugin doesn't support windows

package funplugin

import (
	"fmt"
	"os"
	"testing"

	"github.com/httprunner/funplugin/myexec"
	"github.com/stretchr/testify/assert"
)

func buildGoPlugin() {
	fmt.Println("[setup] build go plugin")
	// flag -race is necessary in order to be consistent with go test
	err := myexec.RunCommand("go", "build", "-buildmode=plugin", "-race",
		"-o=debugtalk.so", "fungo/examples/debugtalk.go")
	if err != nil {
		panic(err)
	}
}

func removeGoPlugin() {
	fmt.Println("[teardown] remove go plugin")
	os.Remove("debugtalk.so")
}

func TestCallPluginFunction(t *testing.T) {
	buildGoPlugin()
	defer removeGoPlugin()

	plugin, err := Init("debugtalk.so", WithDebugLogger(true))
	if err != nil {
		t.Fatal(err)
	}

	if !assert.True(t, plugin.Has("Concatenate")) {
		t.Fail()
	}

	// call function with arguments
	result, err := plugin.Call("Concatenate", "1", 2, "3.14")
	if !assert.NoError(t, err) {
		t.Fail()
	}
	if !assert.Equal(t, "123.14", result) {
		t.Fail()
	}
}
