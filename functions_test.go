package main

import "testing"

func TestNoFuncCollisions(_ *testing.T) {
	_ = tplFuncs.GetFuncMap(nil, nil, nil)
}
