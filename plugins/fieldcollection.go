package plugins

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrValueNotSet   = errors.New("specified value not found")
	ErrValueMismatch = errors.New("specified value has different format")
)

type FieldCollection struct {
	data map[string]interface{}
	lock sync.RWMutex
}

// NewFieldCollection creates a new FieldCollection with empty data store
func NewFieldCollection() *FieldCollection {
	return &FieldCollection{data: make(map[string]interface{})}
}

// FieldCollectionFromData is a wrapper around NewFieldCollection and SetFromData
func FieldCollectionFromData(data map[string]interface{}) *FieldCollection {
	o := NewFieldCollection()
	o.SetFromData(data)
	return o
}

// CanBool tries to read key name as bool and checks whether error is nil
func (f *FieldCollection) CanBool(name string) bool {
	_, err := f.Bool(name)
	return err == nil
}

// CanDuration tries to read key name as time.Duration and checks whether error is nil
func (f *FieldCollection) CanDuration(name string) bool {
	_, err := f.Duration(name)
	return err == nil
}

// CanInt64 tries to read key name as int64 and checks whether error is nil
func (f *FieldCollection) CanInt64(name string) bool {
	_, err := f.Int64(name)
	return err == nil
}

// CanString tries to read key name as string and checks whether error is nil
func (f *FieldCollection) CanString(name string) bool {
	_, err := f.String(name)
	return err == nil
}

// Clone is a wrapper around n.SetFromData(o.Data())
func (f *FieldCollection) Clone() *FieldCollection {
	out := new(FieldCollection)
	out.SetFromData(f.Data())
	return out
}

// Data creates a map-copy of the data stored inside the FieldCollection
func (f *FieldCollection) Data() map[string]interface{} {
	f.lock.RLock()
	defer f.lock.RUnlock()

	out := make(map[string]interface{})
	for k := range f.data {
		out[k] = f.data[k]
	}

	return out
}

// Expect takes a list of keys and returns an error with all non-found names
func (f *FieldCollection) Expect(keys ...string) error {
	f.lock.RLock()
	defer f.lock.RUnlock()

	var missing []string

	for _, k := range keys {
		if _, ok := f.data[k]; !ok {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return errors.Errorf("missing key(s) %s", strings.Join(missing, ", "))
	}

	return nil
}

// HasAll takes a list of keys and returns whether all of them exist inside the FieldCollection
func (f *FieldCollection) HasAll(keys ...string) bool {
	f.lock.RLock()
	defer f.lock.RUnlock()

	for _, k := range keys {
		if _, ok := f.data[k]; !ok {
			return false
		}
	}

	return true
}

// MustBool is a wrapper around Bool and panics if an error was returned
func (f *FieldCollection) MustBool(name string, defVal *bool) bool {
	v, err := f.Bool(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

// MustDuration is a wrapper around Duration and panics if an error was returned
func (f *FieldCollection) MustDuration(name string, defVal *time.Duration) time.Duration {
	v, err := f.Duration(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

// MustInt64 is a wrapper around Int64 and panics if an error was returned
func (f *FieldCollection) MustInt64(name string, defVal *int64) int64 {
	v, err := f.Int64(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

// MustString is a wrapper around String and panics if an error was returned
func (f *FieldCollection) MustString(name string, defVal *string) string {
	v, err := f.String(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

// Bool tries to read key name as bool
func (f *FieldCollection) Bool(name string) (bool, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return false, ErrValueNotSet
	}

	switch v := v.(type) {
	case bool:
		return v, nil
	case string:
		bv, err := strconv.ParseBool(v)
		return bv, errors.Wrap(err, "parsing string to bool")
	}

	return false, ErrValueMismatch
}

// Duration tries to read key name as time.Duration
func (f *FieldCollection) Duration(name string) (time.Duration, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	v, err := f.String(name)
	if err != nil {
		return 0, errors.Wrap(err, "getting string value")
	}

	d, err := time.ParseDuration(v)
	return d, errors.Wrap(err, "parsing value")
}

// Int64 tries to read key name as int64
func (f *FieldCollection) Int64(name string) (int64, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return 0, ErrValueNotSet
	}

	switch v := v.(type) {
	case int:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	}

	return 0, ErrValueMismatch
}

// Set sets a single key to specified value
func (f *FieldCollection) Set(key string, value interface{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.data == nil {
		f.data = make(map[string]interface{})
	}

	f.data[key] = value
}

// SetFromData takes a map of data and copies all data into the FieldCollection
func (f *FieldCollection) SetFromData(data map[string]interface{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.data == nil {
		f.data = make(map[string]interface{})
	}

	for key, value := range data {
		f.data[key] = value
	}
}

// String tries to read key name as string
func (f *FieldCollection) String(name string) (string, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return "", ErrValueNotSet
	}

	if sv, ok := v.(string); ok {
		return sv, nil
	}

	if iv, ok := v.(fmt.Stringer); ok {
		return iv.String(), nil
	}

	return "", ErrValueMismatch
}

// StringSlice tries to read key name as []string
func (f *FieldCollection) StringSlice(name string) ([]string, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return nil, ErrValueNotSet
	}

	switch v := v.(type) {
	case []string:
		return v, nil

	case []interface{}:
		var out []string

		for _, iv := range v {
			sv, ok := iv.(string)
			if !ok {
				return nil, errors.New("value in slice was not string")
			}
			out = append(out, sv)
		}

		return out, nil
	}

	return nil, ErrValueMismatch
}

// Implement JSON marshalling to plain underlying map[string]interface{}

func (f *FieldCollection) MarshalJSON() ([]byte, error) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	return json.Marshal(f.data)
}

func (f *FieldCollection) UnmarshalJSON(raw []byte) error {
	data := make(map[string]interface{})
	if err := json.Unmarshal(raw, &data); err != nil {
		return errors.Wrap(err, "unmarshalling from JSON")
	}

	f.SetFromData(data)
	return nil
}

// Implement YAML marshalling to plain underlying map[string]interface{}

func (f *FieldCollection) MarshalYAML() (interface{}, error) {
	return f.Data(), nil
}

func (f *FieldCollection) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]interface{})
	if err := unmarshal(&data); err != nil {
		return errors.Wrap(err, "unmarshalling from YAML")
	}

	f.SetFromData(data)
	return nil
}
