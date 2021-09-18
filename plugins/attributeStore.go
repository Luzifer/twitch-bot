package plugins

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	errValueNotSet   = errors.New("specified value not found")
	errValueMismatch = errors.New("specified value has different format")
)

type AttributeStore map[string]interface{}

func (a AttributeStore) Expect(keys ...string) error {
	var missing []string

	for _, k := range keys {
		if _, ok := a[k]; !ok {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return errors.Errorf("missing key(s) %s", strings.Join(missing, ", "))
	}

	return nil
}

func (a AttributeStore) MustBool(name string, defVal *bool) bool {
	v, err := a.Bool(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

func (a AttributeStore) MustDuration(name string, defVal *time.Duration) time.Duration {
	v, err := a.Duration(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

func (a AttributeStore) MustInt64(name string, defVal *int64) int64 {
	v, err := a.Int64(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

func (a AttributeStore) MustString(name string, defVal *string) string {
	v, err := a.String(name)
	if err != nil {
		if defVal != nil {
			return *defVal
		}
		panic(err)
	}
	return v
}

func (a AttributeStore) Bool(name string) (bool, error) {
	v, ok := a[name]
	if !ok {
		return false, errValueNotSet
	}

	switch v := v.(type) {
	case bool:
		return v, nil
	case string:
		bv, err := strconv.ParseBool(v)
		return bv, errors.Wrap(err, "parsing string to bool")
	}

	return false, errValueMismatch
}

func (a AttributeStore) Duration(name string) (time.Duration, error) {
	v, err := a.String(name)
	if err != nil {
		return 0, errors.Wrap(err, "getting string value")
	}

	d, err := time.ParseDuration(v)
	return d, errors.Wrap(err, "parsing value")
}

func (a AttributeStore) Int64(name string) (int64, error) {
	v, ok := a[name]
	if !ok {
		return 0, errValueNotSet
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

	return 0, errValueMismatch
}

func (a AttributeStore) String(name string) (string, error) {
	v, ok := a[name]
	if !ok {
		return "", errValueNotSet
	}

	if sv, ok := v.(string); ok {
		return sv, nil
	}

	if iv, ok := v.(fmt.Stringer); ok {
		return iv.String(), nil
	}

	return "", errValueMismatch
}

func (a AttributeStore) StringSlice(name string) ([]string, error) {
	v, ok := a[name]
	if !ok {
		return nil, errValueNotSet
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

	return nil, errValueMismatch
}
