package main

import (
	"strings"
	"sync"
	"text/template"
	"time"

	korvike "github.com/Luzifer/korvike/functions"
	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	log "github.com/sirupsen/logrus"
)

var tplFuncs = newTemplateFuncProvider()

type templateFuncProvider struct {
	funcs map[string]plugins.TemplateFuncGetter
	lock  *sync.RWMutex
}

func newTemplateFuncProvider() *templateFuncProvider {
	out := &templateFuncProvider{
		funcs: map[string]plugins.TemplateFuncGetter{},
		lock:  new(sync.RWMutex),
	}

	return out
}

func (t *templateFuncProvider) GetFuncMap(m *irc.Message, r *plugins.Rule, fields map[string]interface{}) template.FuncMap {
	t.lock.RLock()
	defer t.lock.RUnlock()

	out := make(template.FuncMap)

	for n, fg := range t.funcs {
		out[n] = fg(m, r, fields)
	}

	return out
}

func (t *templateFuncProvider) Register(name string, fg plugins.TemplateFuncGetter) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if _, ok := t.funcs[name]; ok {
		log.Fatalf("Duplicate registration of %q template function", name) //nolint:gocritic // Yeah, the unlock will not run but the process will end
	}

	t.funcs[name] = fg
}

func init() {
	// Register Korvike functions
	for n, f := range korvike.GetFunctionMap() {
		tplFuncs.Register(n, plugins.GenericTemplateFunctionGetter(f))
	}

	tplFuncs.Register("toLower", plugins.GenericTemplateFunctionGetter(strings.ToLower))
	tplFuncs.Register("toUpper", plugins.GenericTemplateFunctionGetter(strings.ToUpper))
	tplFuncs.Register("followDate", plugins.GenericTemplateFunctionGetter(func(from, to string) (time.Time, error) { return twitchClient.GetFollowDate(from, to) }))
	tplFuncs.Register("concat", plugins.GenericTemplateFunctionGetter(func(delim string, parts ...string) string { return strings.Join(parts, delim) }))
	tplFuncs.Register("variable", plugins.GenericTemplateFunctionGetter(func(name string, defVal ...string) string {
		value := store.GetVariable(name)
		if value == "" && len(defVal) > 0 {
			return defVal[0]
		}
		return value
	}))
}
