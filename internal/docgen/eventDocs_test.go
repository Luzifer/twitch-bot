package docgen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Luzifer/twitch-bot/v3/pkg/event"
)

type (
	testCustomEvent struct{}

	testEmptyEvent struct{}

	testTypedEvent struct {
		event.BaseChannel
		Complex  struct{}      `field:"complex" description:"Complex field"`
		Duration time.Duration `field:"duration" description:"Duration field"`
		Names    []string      `field:"names" description:"Names field"`
		Optional *bool         `field:"optional,omitzero" description:"Optional field"`
	}
)

func (testCustomEvent) Description() string { return "Custom description." }
func (testCustomEvent) Event() *string      { return new("custom") }
func (testEmptyEvent) Description() string  { return "Empty description." }
func (testEmptyEvent) Event() *string       { return new("whisper") }
func (testTypedEvent) Description() string  { return "Typed description." }
func (testTypedEvent) Event() *string       { return new("alpha") }

func TestGenerateEventTypes(t *testing.T) {
	content := generateIntoTemporaryWorkingDirectory(t, eventTypesPath, func() error {
		return GenerateEventTypes([]event.DocumentedEvent{testTypedEvent{}, testCustomEvent{}, testEmptyEvent{}})
	})

	assert.Contains(t, content, "duration: number")
	assert.Contains(t, content, "names: Array<string>")
	assert.Contains(t, content, "optional?: boolean")
	assert.Contains(t, content, "complex: unknown")
	assert.Contains(t, content, "SocketMessage<'whisper', null>")
	assert.Contains(t, content, "Record<string, unknown> & { channel: string }")
	assert.Less(t, strings.Index(content, "AlphaFields"), strings.Index(content, "CustomFields"))
}

func TestGenerateEventDocs(t *testing.T) {
	content := generateIntoTemporaryWorkingDirectory(t, eventDocsPath, func() error {
		return GenerateEventDocs([]event.DocumentedEvent{testTypedEvent{}, testCustomEvent{}, testEmptyEvent{}})
	})

	assert.Contains(t, content, "`optional` _*bool_ _(optional)_")
	assert.Contains(t, content, "Any additional fields passed to the custom event")
	assert.Contains(t, content, "This event has no event fields.")
	assert.Less(t, strings.Index(content, "## `alpha`"), strings.Index(content, "## `custom`"))
	assert.Less(t, strings.Index(content, "## `custom`"), strings.Index(content, "## `whisper`"))
}

func TestEventTypeName(t *testing.T) {
	assert.Equal(t, "ChannelpointRedeem", eventTypeName("channelpoint_redeem"))
}

func TestJSType(t *testing.T) {
	assert.Equal(t, "string", jsType(reflect.TypeFor[time.Time]()))
	assert.Equal(t, "number", jsType(reflect.TypeFor[time.Duration]()))
	assert.Equal(t, "Array<number>", jsType(reflect.TypeFor[[]int64]()))
}

func TestEventClientDeclarationConsumer(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)

	tempDir := t.TempDir()

	//#nosec:G304 // read of generated file inside code-tree
	eventClient, err := os.ReadFile(filepath.Join(repoRoot, "internal/apimodules/overlays/src/eventclient.ts"))
	require.NoError(t, err)

	//#nosec:G304 // read of generated file inside code-tree
	usage, err := os.ReadFile(filepath.Join(repoRoot, "internal/docgen/testdata/usage.ts"))
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "eventclient.ts"), eventClient, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "usage.ts"), usage, 0o600))

	t.Chdir(tempDir)
	require.NoError(t, os.MkdirAll(filepath.Dir(eventTypesPath), 0o750))
	require.NoError(t, GenerateEventTypes([]event.DocumentedEvent{testTypedEvent{}, testCustomEvent{}, testEmptyEvent{}}))
	require.NoError(t, os.Rename(eventTypesPath, "eventTypes.ts"))

	cmd := exec.CommandContext(
		context.TODO(),
		//#nosec:G204 -- executable path is rooted in the checked-out repository
		filepath.Join(repoRoot, "node_modules/.bin/tsc"),
		"--ignoreConfig",
		"--lib", "DOM,ES2020",
		"--module", "ESNext",
		"--moduleResolution", "Bundler",
		"--noEmit",
		"--strict",
		"--target", "ES2020",
		"usage.ts",
	)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}

func generateIntoTemporaryWorkingDirectory(t *testing.T, path string, generate func() error) string {
	t.Helper()

	t.Chdir(t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, generate())

	//#nosec:G304 // path is a generator output constant supplied by the test
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}
