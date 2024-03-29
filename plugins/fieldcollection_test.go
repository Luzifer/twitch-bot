package plugins

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFieldCollectionJSONMarshal(t *testing.T) {
	var (
		buf = new(bytes.Buffer)
		raw = `{"key1":"test1","key2":"test2"}`
		f   = NewFieldCollection()
	)

	if err := json.NewDecoder(strings.NewReader(raw)).Decode(f); err != nil {
		t.Fatalf("Unable to unmarshal: %s", err)
	}

	if err := json.NewEncoder(buf).Encode(f); err != nil {
		t.Fatalf("Unable to marshal: %s", err)
	}

	if raw != strings.TrimSpace(buf.String()) {
		t.Errorf("Marshalled JSON does not match expectation: res=%s exp=%s", buf.String(), raw)
	}
}

func TestFieldCollectionYAMLMarshal(t *testing.T) {
	var (
		buf = new(bytes.Buffer)
		raw = "key1: test1\nkey2: test2"
		f   = NewFieldCollection()
	)

	if err := yaml.NewDecoder(strings.NewReader(raw)).Decode(f); err != nil {
		t.Fatalf("Unable to unmarshal: %s", err)
	}

	if err := yaml.NewEncoder(buf).Encode(f); err != nil {
		t.Fatalf("Unable to marshal: %s", err)
	}

	if raw != strings.TrimSpace(buf.String()) {
		t.Errorf("Marshalled YAML does not match expectation: res=%s exp=%s", buf.String(), raw)
	}
}

func TestFieldCollectionNilModify(_ *testing.T) {
	var f *FieldCollection

	f.Set("foo", "bar")

	f = nil
	f.SetFromData(map[string]interface{}{"foo": "bar"})
}

func TestFieldCollectionNilClone(_ *testing.T) {
	var f *FieldCollection

	f.Clone()
}

func TestFieldCollectionNilDataGet(t *testing.T) {
	var f *FieldCollection

	for name, fn := range map[string]func(name string) bool{
		"bool":     f.CanBool,
		"duration": f.CanDuration,
		"int64":    f.CanInt64,
		"string":   f.CanString,
	} {
		if fn("foo") {
			t.Errorf("%s key is available", name)
		}
	}
}

func TestFieldCollectionIntToString(t *testing.T) {
	val := 123
	fc := FieldCollectionFromData(map[string]interface{}{"test": val})

	if !fc.CanString("test") {
		t.Fatalf("cannot convert %T to string", val)
	}

	if v := fc.MustString("test", nil); v != "123" {
		t.Errorf("unexpected value: 123 != %s", v)
	}
}
