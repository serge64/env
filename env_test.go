package env_test

import (
	"os"
	"testing"
	"time"

	"github.com/serge64/env"
)

type ValidStruct struct {
	// Home should match Environ because it has a "env" field tag.
	Home string `env:"HOME"`

	// Jenkins should be recursed into.
	Jenkins struct {
		Workspace string `env:"WORKSPACE"`

		// PointerMissing should not be set if the environment variable is missing.
		PointerMissing *string `env:"JENKINS_POINTER_MISSING"`
	}

	// PointerString should be nil if unset, with "" being a valid value.
	PointerString *string `env:"POINTER_STRING"`

	// PointerInt should work along with other supported types.
	PointerInt *int `env:"POINTER_INT"`

	// PointerPointerString should be recursed into.
	PointerPointerString **string `env:"POINTER_POINTER_STRING"`

	// PointerMissing should not be set if the environment variable is missing.
	PointerMissing *string `env:"POINTER_MISSING"`

	// Extra should remain with a zero-value because it has no "env" field tag.
	Extra string

	// Additional supported types
	Int     int     `env:"INT"`
	Float32 float32 `env:"FLOAT32"`
	Float64 float64 `env:"FLOAT64"`
	Bool    bool    `env:"BOOL"`

	// time.Duration is supported
	Duration time.Duration `env:"TYPE_DURATION"`
}

type UnsupportedStruct struct {
	Timestamp time.Time `env:"TIMESTAMP"`
}

type UnexportedStruct struct {
	home string `env:"HOME"`
}

type DefaultValueStruct struct {
	DefaultString             string        `env:"MISSING_STRING,default=found"`
	DefaultKeyValueString     string        `env:"MISSING_KVSTRING,default=key=value"`
	DefaultBool               bool          `env:"MISSING_BOOL,default=true"`
	DefaultInt                int           `env:"MISSING_INT,default=7"`
	DefaultFloat32            float32       `env:"MISSING_FLOAT32,default=8.9"`
	DefaultFloat64            float64       `env:"MISSING_FLOAT64,default=10.11"`
	DefaultDuration           time.Duration `env:"MISSING_DURATION,default=5s"`
	DefaultWithOptionsMissing string        `env:"MISSING_1,MISSING_2,default=present"`
	DefaultWithOptionsPresent string        `env:"MISSING_1,PRESENT,default=present"`
}

func TestUnmarshal(t *testing.T) {
	environ := map[string]string{
		"HOME":             "/home/test",
		"WORKSPACE":        "/mnt/builds/slave/workspace/test",
		"EXTRA":            "extra",
		"INT":              "1",
		"FLOAT32":          "2.3",
		"FLOAT64":          "4.5",
		"BOOL":             "true",
		"NPM_CONFIG_CACHE": "second",
		"TYPE_DURATION":    "5s",
	}

	for k, v := range environ {
		_ = os.Setenv(k, v)
	}

	var validStruct ValidStruct
	err := env.Unmarshal(&validStruct)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if validStruct.Home != "/home/test" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "/home/test", validStruct.Home)
	}

	if validStruct.Jenkins.Workspace != "/mnt/builds/slave/workspace/test" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "/mnt/builds/slave/workspace/test", validStruct.Jenkins.Workspace)
	}

	if validStruct.PointerString != nil {
		t.Errorf("Expected field value to be '%v' but got '%v'", nil, validStruct.PointerString)
	}

	if validStruct.Extra != "" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "", validStruct.Extra)
	}

	if validStruct.Int != 1 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 1, validStruct.Int)
	}

	if validStruct.Float32 != 2.3 {
		t.Errorf("Expected field value to be '%f' but got '%f'", 2.3, validStruct.Float32)
	}

	if validStruct.Float64 != 4.5 {
		t.Errorf("Expected field value to be '%f' but got '%f'", 4.5, validStruct.Float64)
	}

	if validStruct.Bool != true {
		t.Errorf("Expected field value to be '%t' but got '%t'", true, validStruct.Bool)
	}

	if validStruct.Duration != 5*time.Second {
		t.Errorf("Expected field value to be '%s' but got '%s'", "5s", validStruct.Duration)
	}
}

func TestUnmarshalPointer(t *testing.T) {
	environ := map[string]string{
		"POINTER_STRING":         "",
		"POINTER_INT":            "1",
		"POINTER_POINTER_STRING": "",
	}

	for k, v := range environ {
		_ = os.Setenv(k, v)
	}

	var validStruct ValidStruct
	err := env.Unmarshal(&validStruct)
	if err != nil {
		t.Errorf("Expected no error but got '%s'", err)
	}

	if validStruct.PointerString == nil {
		t.Errorf("Expected field value to be '%s' but got '%v'", "", nil)
	} else if *validStruct.PointerString != "" {
		t.Errorf("Expected field value to be '%s' but got '%s'", "", *validStruct.PointerString)
	}

	if validStruct.PointerInt == nil {
		t.Errorf("Expected field value to be '%d' but got '%v'", 1, nil)
	} else if *validStruct.PointerInt != 1 {
		t.Errorf("Expected field value to be '%d' but got '%d'", 1, *validStruct.PointerInt)
	}

	if validStruct.PointerPointerString == nil {
		t.Errorf("Expected field value to be '%s' but got '%v'", "", nil)
	} else {
		if *validStruct.PointerPointerString == nil {
			t.Errorf("Expected field value to be '%s' but got '%v'", "", nil)
		} else if **validStruct.PointerPointerString != "" {
			t.Errorf("Expected field value to be '%s' but got '%s'", "", **validStruct.PointerPointerString)
		}
	}

	if validStruct.PointerMissing != nil {
		t.Errorf("Expected field value to be '%v' but got '%s'", nil, *validStruct.PointerMissing)
	}
}

func TestUnmarshalInvalid(t *testing.T) {
	var validStruct ValidStruct
	err := env.Unmarshal(validStruct)
	if err != env.ErrInvalidValue {
		t.Errorf("Expected error 'ErrInvalidValue' but got '%s'", err)
	}

	ptr := &validStruct
	err = env.Unmarshal(&ptr)
	if err != env.ErrInvalidValue {
		t.Errorf("Expected error 'ErrInvalidValue' but got '%s'", err)
	}
}

func TestUnmarshalUnsupported(t *testing.T) {
	_ = os.Setenv("TIMESTAMP", "2016-07-15T12:00:00.000Z")

	var unsupportedStruct UnsupportedStruct
	err := env.Unmarshal(&unsupportedStruct)
	if err != env.ErrUnsupportedType {
		t.Errorf("Expected error 'ErrUnsupportedType' but got '%s'", err)
	}
}

func TestUnmarshalUnexported(t *testing.T) {
	_ = os.Setenv("HOME", "/home/test")

	var unexportedStruct UnexportedStruct
	err := env.Unmarshal(&unexportedStruct)
	if err != env.ErrUnexportedField {
		t.Errorf("Expected error 'ErrUnexportedField' but got '%s'", err)
	}

	if unexportedStruct.home != "" {
		t.Errorf("Expected empty value but got '%s'", unexportedStruct.home)
	}
}

func TestUnmarshalDefaultValues(t *testing.T) {
	_ = os.Setenv("PRESENT", "youFoundMe")

	var defaultValueStruct DefaultValueStruct
	err := env.Unmarshal(&defaultValueStruct)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	testCases := [][]interface{}{
		{defaultValueStruct.DefaultInt, 7},
		{defaultValueStruct.DefaultFloat32, float32(8.9)},
		{defaultValueStruct.DefaultFloat64, 10.11},
		{defaultValueStruct.DefaultBool, true},
		{defaultValueStruct.DefaultString, "found"},
		{defaultValueStruct.DefaultKeyValueString, "key=value"},
		{defaultValueStruct.DefaultDuration, 5 * time.Second},
		{defaultValueStruct.DefaultWithOptionsMissing, "present"},
		{defaultValueStruct.DefaultWithOptionsPresent, "youFoundMe"},
	}

	for _, testCase := range testCases {
		if testCase[0] != testCase[1] {
			t.Errorf("Expected field value to be '%v' but got '%v'", testCase[1], testCase[0])
		}
	}
}
