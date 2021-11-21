package env

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrInvalidValue returned when the value passed to Unmarshal is nil or not a
	// pointer to a struct.
	ErrInvalidValue = errors.New("value must be a non-nil pointer to a structure")

	// ErrUnexportedField returned when a field with tag "env" is not exported.
	ErrUnexportedField = errors.New("field must be exported")

	// ErrUnsupportedType returned when a field with tag "env" is unsupported.
	ErrUnsupportedType = errors.New("field is an unsupported type")
)

type envSet map[string]string

type tag struct {
	Key     string
	Default string
}

// Unmarshal parses os.Environ and stores the result at the value
// pointed to by v.
//
// If v is zero or not a pointer to a structure, Unmarshal returns
// ErrInvalidValue.
//
// If fields tagged with "env" are not exported, Unmarshal returns
// ErrUnexportedField.
//
// If the field is of an unsupported type, Unmarshal returns
// ErrUnsupportedType.
func Unmarshal(v interface{}) error {
	es := environToEnvSet(os.Environ())
	return unmarshal(es, v)
}

func environToEnvSet(environ []string) envSet {
	m := make(envSet, len(environ))
	for _, v := range environ {
		parts := strings.SplitN(v, "=", 2)
		m[parts[0]] = parts[1]
	}
	return m
}

func unmarshal(es envSet, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrInvalidValue
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return ErrInvalidValue
	}

	t := rv.Type()

	for i := 0; i < t.NumField(); i++ {
		valueField := rv.Field(i)
		switch valueField.Kind() {
		case reflect.Struct:
			if !valueField.Addr().CanInterface() {
				continue
			}

			iface := valueField.Addr().Interface()
			err := unmarshal(es, iface)
			if err != nil {
				return err
			}
		}

		typeField := t.Field(i)
		tag := typeField.Tag.Get("env")
		if tag == "" {
			continue
		}

		if !valueField.CanSet() {
			return ErrUnexportedField
		}

		envTag := parseTag(tag)

		envValue, ok := es[envTag.Key]
		if !ok {
			if envTag.Default == "" {
				continue
			} else {
				envValue = envTag.Default
			}
		}

		err := set(typeField.Type, valueField, envValue)
		if err != nil {
			return err
		}

		delete(es, tag)
	}

	return nil
}

func parseTag(tagString string) tag {
	var t tag
	envKeys := strings.Split(tagString, ",")
	for _, key := range envKeys {
		if strings.Contains(key, "=") {
			keyData := strings.SplitN(key, "=", 2)
			if strings.ToLower(keyData[0]) == "default" {
				t.Default = keyData[1]
			}
			continue
		}
		t.Key = key
	}
	return t
}

func set(t reflect.Type, f reflect.Value, value string) error {
	switch t.Kind() {
	case reflect.Ptr:
		ptr := reflect.New(t.Elem())
		err := set(t.Elem(), ptr.Elem(), value)
		if err != nil {
			return err
		}
		f.Set(ptr)
	case reflect.String:
		f.SetString(value)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		f.SetBool(v)
	case reflect.Float32:
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return err
		}
		f.SetFloat(v)
	case reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.SetFloat(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if t.PkgPath() == "time" && t.Name() == "Duration" {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(duration))
			break
		}
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		f.SetInt(int64(v))
	default:
		return ErrUnsupportedType
	}
	return nil
}
