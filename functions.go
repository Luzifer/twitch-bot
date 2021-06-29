package main

import (
	"strings"
	"sync"
	"text/template"

	korvike "github.com/Luzifer/korvike/functions"
	"github.com/go-irc/irc"
)

var tplFuncs = newTemplateFuncProvider()

type (
	templateFuncGetter   func(*irc.Message, *Rule, map[string]interface{}) interface{}
	templateFuncProvider struct {
		funcs map[string]templateFuncGetter
		lock  *sync.RWMutex
	}
)

func newTemplateFuncProvider() *templateFuncProvider {
	out := &templateFuncProvider{
		funcs: map[string]templateFuncGetter{},
		lock:  new(sync.RWMutex),
	}

	return out
}

func (t *templateFuncProvider) GetFuncMap(m *irc.Message, r *Rule, fields map[string]interface{}) template.FuncMap {
	t.lock.RLock()
	defer t.lock.RUnlock()

	out := make(template.FuncMap)

	for n, fg := range t.funcs {
		out[n] = fg(m, r, fields)
	}

	return out
}

func (t *templateFuncProvider) Register(name string, fg templateFuncGetter) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.funcs[name] = fg
}

func genericTemplateFunctionGetter(f interface{}) templateFuncGetter {
	return func(*irc.Message, *Rule, map[string]interface{}) interface{} { return f }
}

func init() {
	// Register Korvike functions
	for n, f := range korvike.GetFunctionMap() {
		tplFuncs.Register(n, genericTemplateFunctionGetter(f))
	}

	tplFuncs.Register("toLower", genericTemplateFunctionGetter(strings.ToLower))
	tplFuncs.Register("toUpper", genericTemplateFunctionGetter(strings.ToUpper))
	tplFuncs.Register("followDate", genericTemplateFunctionGetter(twitch.GetFollowDate))
	tplFuncs.Register("concat", genericTemplateFunctionGetter(func(delim string, parts ...string) string { return strings.Join(parts, delim) }))
	tplFuncs.Register("variable", genericTemplateFunctionGetter(func(name string, defVal ...string) string {
		value := store.GetVariable(name)
		if value == "" && len(defVal) > 0 {
			return defVal[0]
		}
		return value
	}))
}
