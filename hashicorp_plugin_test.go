package funplugin

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func buildHashicorpGoPlugin() {
	log.Info().Msg("[init] build hashicorp go plugin")
	cmd := exec.Command("go", "build",
		"-o", "fungo/examples/debugtalk.bin",
		"fungo/examples/hashicorp.go", "fungo/examples/debugtalk.go")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func removeHashicorpGoPlugin() {
	log.Info().Msg("[teardown] remove hashicorp plugin")
	os.Remove("fungo/examples/debugtalk.bin")
}

func TestHashicorpGoPlugin(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	plugin, err := Init("fungo/examples/debugtalk.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer plugin.Quit()

	assertPlugin(t, plugin)
}

func TestHashicorpPythonPluginWithoutVenv(t *testing.T) {
	_, err := Init("funppy/examples/debugtalk.py")
	assert.EqualError(t, err, "python3 not specified")
}

func TestHashicorpPythonPluginWithVenv(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "prefix")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	venvDir := filepath.Join(dir, ".hrp", "venv")
	err = exec.Command("python3", "-m", "venv", venvDir).Run()
	if err != nil {
		t.Fatal(err)
	}

	var python3 string
	if runtime.GOOS == "windows" {
		python3 = filepath.Join(venvDir, "Scripts", "python3.exe")
	} else {
		python3 = filepath.Join(venvDir, "bin", "python3")
	}

	err = exec.Command(python3, "-m", "pip", "install", "funppy").Run()
	if err != nil {
		t.Fatal(err)
	}

	plugin, err := Init("funppy/examples/debugtalk.py", WithPython3(python3))
	if err != nil {
		t.Fatal(err)
	}
	defer plugin.Quit()

	assertPlugin(t, plugin)
}

func assertPlugin(t *testing.T, plugin IPlugin) {
	var err error
	if !assert.True(t, plugin.Has("sum_ints")) {
		t.Fail()
	}
	if !assert.True(t, plugin.Has("concatenate")) {
		t.Fail()
	}

	var v2 interface{}
	v2, err = plugin.Call("sum_ints", 1, 2, 3, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.EqualValues(t, 10, v2) {
		t.Fail()
	}
	v2, err = plugin.Call("sum_two_int", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.EqualValues(t, 3, v2) {
		t.Fail()
	}
	v2, err = plugin.Call("sum", 1, 2, 3.4, 5)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, 11.4, v2) {
		t.Fail()
	}

	var v3 interface{}
	v3, err = plugin.Call("sum_two_string", "a", "b")
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, "ab", v3) {
		t.Fail()
	}
	v3, err = plugin.Call("sum_strings", "a", "b", "c")
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, "abc", v3) {
		t.Fail()
	}

	v3, err = plugin.Call("concatenate", "a", 2, "c", 3.4)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, "a2c3.4", v3) {
		t.Fail()
	}
}
