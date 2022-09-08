package main

import "github.com/Luzifer/twitch-bot/plugins"

func init() {
	tplFuncs.Register("variable", plugins.GenericTemplateFunctionGetter(func(name string, defVal ...string) string {
		value := store.GetVariable(name)
		if value == "" && len(defVal) > 0 {
			return defVal[0]
		}
		return value
	}))
}
