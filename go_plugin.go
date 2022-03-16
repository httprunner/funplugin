package funplugin

import (
	"fmt"
	"plugin"
	"reflect"
	"runtime"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/funplugin/shared"
)

// goPlugin implements golang official plugin
type goPlugin struct {
	*plugin.Plugin
	cachedFunctions map[string]reflect.Value // cache loaded functions to improve performance
}

func (p *goPlugin) Init(path string) error {
	if runtime.GOOS == "windows" {
		log.Warn().Msg("go plugin does not support windows")
		return fmt.Errorf("go plugin does not support windows")
	}

	var err error
	p.Plugin, err = plugin.Open(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("load go plugin failed")
		return err
	}

	p.cachedFunctions = make(map[string]reflect.Value)
	log.Info().Str("path", path).Msg("load go plugin success")
	return nil
}

func (p *goPlugin) Type() string {
	return "go-plugin"
}

func (p *goPlugin) Has(funcName string) bool {
	fn, ok := p.cachedFunctions[funcName]
	if ok {
		return fn.IsValid()
	}

	sym, err := p.Plugin.Lookup(funcName)
	if err != nil {
		p.cachedFunctions[funcName] = reflect.Value{} // mark as invalid
		return false
	}
	fn = reflect.ValueOf(sym)

	// check function type
	if fn.Kind() != reflect.Func {
		p.cachedFunctions[funcName] = reflect.Value{} // mark as invalid
		return false
	}

	p.cachedFunctions[funcName] = fn
	return true
}

func (p *goPlugin) Call(funcName string, args ...interface{}) (interface{}, error) {
	if !p.Has(funcName) {
		return nil, fmt.Errorf("function %s not found", funcName)
	}
	fn := p.cachedFunctions[funcName]
	return shared.CallFunc(fn, args...)
}

func (p *goPlugin) Quit() error {
	// no need to quit for go plugin
	return nil
}
