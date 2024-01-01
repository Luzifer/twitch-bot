package main

import (
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/v2/str"
	korvike "github.com/Luzifer/korvike/functions"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	korvikeBlacklist = []string{"now"}
	sprigBlacklist   = []string{"env"}
	tplFuncs         = newTemplateFuncProvider()
)

type templateFuncProvider struct {
	docs  []plugins.TemplateFuncDocumentation
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

func (t *templateFuncProvider) Register(name string, fg plugins.TemplateFuncGetter, doc ...plugins.TemplateFuncDocumentation) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if _, ok := t.funcs[name]; ok {
		logrus.Fatalf("Duplicate registration of %q template function", name)
	}

	t.funcs[name] = fg

	if len(doc) > 0 {
		doc[0].Name = name
		t.docs = append(t.docs, doc[0])
	}
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
			dLeft -= part * div //nolint:durationcheck // One is static, this is fine

			if len(units) <= idx || units[idx] == "" {
				continue
			}

			parts = append(parts, fmt.Sprintf("%d %s", part, units[idx]))
		}

		return strings.Join(parts, ", ")
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns a formated duration. Pass empty strings to leave out the specific duration part.",
		Syntax:      "formatDuration <duration> <hours> <minutes> <seconds>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ formatDuration .testDuration "hours" "minutes" "seconds" }} - {{ formatDuration .testDuration "hours" "minutes" "" }}`,
			ExpectedOutput: "5 hours, 33 minutes, 12 seconds - 5 hours, 33 minutes",
		},
	})
}
