package main

import (
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-irc/irc"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/str"
	korvike "github.com/Luzifer/korvike/functions"
	"github.com/Luzifer/twitch-bot/plugins"
)

var (
	korvikeBlacklist = []string{"now"}
	sprigBlacklist   = []string{"env"}
	tplFuncs         = newTemplateFuncProvider()
)

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

func (t *templateFuncProvider) GetFuncMap(m *irc.Message, r *plugins.Rule, fields *plugins.FieldCollection) template.FuncMap {
	t.lock.RLock()
	defer t.lock.RUnlock()

	out := make(template.FuncMap)

	for n, fn := range sprig.TxtFuncMap() {
		if str.StringInSlice(n, sprigBlacklist) {
			continue
		}
		if out[n] != nil {
			panic(fmt.Sprintf("duplicate function: %s (add in sprig)", n))
		}
		out[n] = fn
	}

	for n, fg := range t.funcs {
		if out[n] != nil {
			panic(fmt.Sprintf("duplicate function: %s (add in registration)", n))
		}
		out[n] = fg(m, r, fields)
	}

	return out
}

func (t *templateFuncProvider) GetFuncNames() []string {
	var out []string

	for n := range t.GetFuncMap(nil, nil, nil) {
		out = append(out, n)
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
		if str.StringInSlice(n, korvikeBlacklist) {
			continue
		}
		tplFuncs.Register(n, plugins.GenericTemplateFunctionGetter(f))
	}

	tplFuncs.Register("formatDuration", plugins.GenericTemplateFunctionGetter(func(dur time.Duration, units ...string) string {
		dLeft := dur

		if len(units) == 0 {
			return ""
		}

		var parts []string
		for idx, div := range []time.Duration{time.Hour, time.Minute, time.Second} {
			part := dLeft / div
			dLeft -= part * div

			if len(units) <= idx || units[idx] == "" {
				continue
			}

			parts = append(parts, fmt.Sprintf("%d %s", part, units[idx]))
		}

		return strings.Join(parts, ", ")
	}))
}
