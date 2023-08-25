package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateFuncDocs(t *testing.T) {
	for _, fd := range tplFuncs.docs {
		t.Run(fd.Name, func(t *testing.T) {
			if fd.Example == nil {
				t.Skip("no example present")
			}

			if fd.Example.ExpectedOutput == "" {
				t.Skip("no expected output present")
			}

			out, err := generateTplDocsRender(fd.Example)
			assert.NoError(t, err)
			assert.Equal(t, fd.Example.ExpectedOutput, out)
		})
	}
}
