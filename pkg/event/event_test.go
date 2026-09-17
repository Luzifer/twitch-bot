package event

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	customZero struct {
		Zero bool
	}

	pointerEvent struct {
		Value *string `field:"value,omitzero"`
	}

	pointerZero struct {
		Zero bool
	}

	testEvent struct {
		BaseChannel
		Ignored string `field:"-"`
		Omitted string `field:"omitted,omitzero"`
		Secret  string `field:"secret" nolog:"true"`
		Value   string `field:"value"`
	}
)

func (pointerEvent) Event() *string { return new("pointer") }

func (testEvent) Event() *string { return new("test") }

func (c customZero) IsZero() bool { return c.Zero }

func (p *pointerZero) IsZero() bool { return p.Zero }

func TestIsZero(t *testing.T) {
	var nilInterface any

	pointerReceiver := pointerZero{Zero: true}
	typedNil := any((*pointerZero)(nil))
	interfaceValue := reflect.ValueOf(&typedNil).Elem()

	testCases := map[string]struct {
		Expected bool
		Value    reflect.Value
	}{
		"custom method false overrides zero value": {
			Expected: false,
			Value:    reflect.ValueOf(customZero{}),
		},
		"custom method true overrides non-zero value": {
			Expected: true,
			Value:    reflect.ValueOf(customZero{Zero: true}),
		},
		"nil interface": {
			Expected: true,
			Value:    reflect.ValueOf(&nilInterface).Elem(),
		},
		"nil pointer": {
			Expected: true,
			Value:    reflect.ValueOf((*string)(nil)),
		},
		"non-zero string": {
			Expected: false,
			Value:    reflect.ValueOf("value"),
		},
		"pointer receiver": {
			Expected: true,
			Value:    reflect.ValueOf(&pointerReceiver).Elem(),
		},
		"typed nil in interface": {
			Expected: true,
			Value:    interfaceValue,
		},
		"zero string": {
			Expected: true,
			Value:    reflect.ValueOf(""),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.Expected, isZero(testCase.Value))
		})
	}
}

func TestToFieldCollection(t *testing.T) {
	evt := &testEvent{
		Channel: "#example",
		Ignored: "ignored",
		Secret:  "secret",
		Value:   "value",
	}

	fields := ToFieldCollection(evt)
	require.NotNil(t, fields)
	assert.Equal(t, map[string]any{
		"channel": "#example",
		"secret":  "secret",
		"value":   "value",
	}, fields.Data())
}

func TestToFieldCollectionIncludesNonZeroField(t *testing.T) {
	fields := ToFieldCollection(testEvent{Omitted: "included"})
	require.NotNil(t, fields)
	assert.Equal(t, "included", fields.Data()["omitted"])
}

func TestToFieldCollectionUnwrapsPointer(t *testing.T) {
	value := "value"
	fields := ToFieldCollection(pointerEvent{Value: &value})
	require.NotNil(t, fields)
	assert.Equal(t, "value", fields.Data()["value"])
}

func TestToFieldCollectionWithoutFields(t *testing.T) {
	assert.Nil(t, ToFieldCollection(Whisper{}))
}

func TestToLogFields(t *testing.T) {
	evt := testEvent{
		Channel: "#example",
		Ignored: "ignored",
		Secret:  "secret",
		Value:   "value",
	}

	assert.Equal(t, map[string]any{
		"channel": "#example",
		"value":   "value",
	}, map[string]any(ToLogFields(evt)))
}
