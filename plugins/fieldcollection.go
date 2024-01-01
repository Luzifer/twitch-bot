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
	// ErrValueNotSet is used to notify the value is not available in the FieldCollection
	ErrValueNotSet = errors.New("specified value not found")
	// ErrValueMismatch is used to notify the value does not match the requested type
	ErrValueMismatch = errors.New("specified value has different format")
)

// FieldCollection holds an map[string]any with conversion functions attached
type FieldCollection struct {
	data map[string]any
	lock sync.RWMutex
}

// NewFieldCollection creates a new FieldCollection with empty data store
func NewFieldCollection() *FieldCollection {
	return &FieldCollection{data: make(map[string]any)}
}

// FieldCollectionFromData is a wrapper around NewFieldCollection and SetFromData
func FieldCollectionFromData(data map[string]any) *FieldCollection {
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
func (f *FieldCollection) Data() map[string]any {
	if f == nil {
		return nil
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	out := make(map[string]any)
	for k := range f.data {
		out[k] = f.data[k]
	}

	return out
}

// Expect takes a list of keys and returns an error with all non-found names
func (f *FieldCollection) Expect(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if f == nil || f.data == nil {
		return errors.New("uninitialized field collection")
	}

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
	return f.Expect(keys...) == nil
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

// MustStringSlice is a wrapper around StringSlice and returns nil in case name is not set
func (f *FieldCollection) MustStringSlice(name string) []string {
	v, err := f.StringSlice(name)
	if err != nil {
		return nil
	}
	return v
}

// Any tries to read key name as any-type (interface)
func (f *FieldCollection) Any(name string) (any, error) {
	if f == nil || f.data == nil {
		return false, errors.New("uninitialized field collection")
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return false, ErrValueNotSet
	}

	return v, nil
}

// Bool tries to read key name as bool
func (f *FieldCollection) Bool(name string) (bool, error) {
	if f == nil || f.data == nil {
		return false, errors.New("uninitialized field collection")
	}

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
	if f == nil || f.data == nil {
		return 0, errors.New("uninitialized field collection")
	}

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
	if f == nil || f.data == nil {
		return 0, errors.New("uninitialized field collection")
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return 0, ErrValueNotSet
	}

	switch v := v.(type) {
	case float64:
		return int64(v), nil
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
func (f *FieldCollection) Set(key string, value any) {
	if f == nil {
		f = NewFieldCollection()
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	if f.data == nil {
		f.data = make(map[string]any)
	}

	f.data[key] = value
}

// SetFromData takes a map of data and copies all data into the FieldCollection
func (f *FieldCollection) SetFromData(data map[string]any) {
	if f == nil {
		f = NewFieldCollection()
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	if f.data == nil {
		f.data = make(map[string]any)
	}

	for key, value := range data {
		f.data[key] = value
	}
}

// String tries to read key name as string
func (f *FieldCollection) String(name string) (string, error) {
	if f == nil || f.data == nil {
		return "", errors.New("uninitialized field collection")
	}

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

	return fmt.Sprintf("%v", v), nil
}

// StringSlice tries to read key name as []string
func (f *FieldCollection) StringSlice(name string) ([]string, error) {
	if f == nil || f.data == nil {
		return nil, errors.New("uninitialized field collection")
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	v, ok := f.data[name]
	if !ok {
		return nil, ErrValueNotSet
	}

	switch v := v.(type) {
	case []string:
		return v, nil

	case []any:
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

// Implement JSON marshalling to plain underlying map[string]any

// MarshalJSON implements the json.Marshaller interface
func (f *FieldCollection) MarshalJSON() ([]byte, error) {
	if f == nil || f.data == nil {
		return []byte("{}"), nil
	}

	f.lock.RLock()
	defer f.lock.RUnlock()

	data, err := json.Marshal(f.data)
	if err != nil {
		return nil, fmt.Errorf("marshalling data to json: %w", err)
	}

	return data, nil
}

// UnmarshalJSON implements the json.Unmarshaller interface
func (f *FieldCollection) UnmarshalJSON(raw []byte) error {
	data := make(map[string]any)
	if err := json.Unmarshal(raw, &data); err != nil {
		return errors.Wrap(err, "unmarshalling from JSON")
	}

	f.SetFromData(data)
	return nil
}

// Implement YAML marshalling to plain underlying map[string]any

// MarshalYAML implements the yaml.Marshaller interface
func (f *FieldCollection) MarshalYAML() (any, error) {
	return f.Data(), nil
}

// UnmarshalYAML implements the yaml.Unmarshaller interface
func (f *FieldCollection) UnmarshalYAML(unmarshal func(any) error) error {
	data := make(map[string]any)
	if err := unmarshal(&data); err != nil {
		return errors.Wrap(err, "unmarshalling from YAML")
	}

	f.SetFromData(data)
	return nil
}
