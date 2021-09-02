package plugins

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrValueNotSet   = errors.New("specified value not found")
	ErrValueMismatch = errors.New("specified value has different format")
)

type FieldCollection map[string]interface{}

func (m FieldCollection) Expect(keys ...string) error {
	var missing []string

	for _, k := range keys {
		if _, ok := m[k]; !ok {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return errors.Errorf("missing key(s) %s", strings.Join(missing, ", "))
	}

	return nil
}

func (f FieldCollection) MustBool(name string, defVal *bool) bool {
	v, err := f.Bool(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

func (f FieldCollection) MustDuration(name string, defVal *time.Duration) time.Duration {
	v, err := f.Duration(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

func (f FieldCollection) MustInt64(name string, defVal *int64) int64 {
	v, err := f.Int64(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

func (f FieldCollection) MustString(name string, defVal *string) string {
	v, err := f.String(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

func (f FieldCollection) Bool(name string) (bool, error) {
	v, ok := f[name]
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

func (f FieldCollection) Duration(name string) (time.Duration, error) {
	v, err := f.String(name)
	if err != nil {
		return 0, errors.Wrap(err, "getting string value")
	}

	d, err := time.ParseDuration(v)
	return d, errors.Wrap(err, "parsing value")
}

func (f FieldCollection) Int64(name string) (int64, error) {
	v, ok := f[name]
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

func (f FieldCollection) String(name string) (string, error) {
	v, ok := f[name]
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

func (f FieldCollection) StringSlice(name string) ([]string, error) {
	v, ok := f[name]
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
