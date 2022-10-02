package main

import "testing"

func TestNoFuncCollisions(t *testing.T) {
	_ = tplFuncs.GetFuncMap(nil, nil, nil)
}
