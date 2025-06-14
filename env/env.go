package env

import (
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// rex is a compiled regular expression to match pattern like ${templateFile:variableName}
var rex = regexp.MustCompile(`\${([^:]+):([^}]+)}`)

type VarReader interface {
	// GetString returns the value associated with the key as a string
	GetString(key string) string

	// IsSet checks to see if the key has been set in any of the data locations.
	IsSet(key string) bool
}

// ModuleConfig is used to read config
type ModuleConfig interface {
	// SubReader returns new ModuleConfig instance representing a sub tree of this instance.
	// SubReader is case-insensitive for a key.
	SubConfig(key string) ModuleConfig

	// GetString returns the value associated with the key as a string
	GetString(key string) string

	// GetBool returns the value associated with the key as a boolean.
	GetBool(key string) bool

	// GetInt64 returns the value associated with the key as an integer.
	GetInt64(key string) int64

	// GetInt returns the value associated with the key as an integer.
	GetInt(key string) int

	// GetFloat64 returns the value associated with the key as a float64.
	GetFloat64(key string) float64

	// GetTime returns the value associated with the key as time.
	GetTime(key string) time.Time

	// GetStringMapString returns the value associated with the key as a map of strings.
	GetStringMapString(key string) map[string]string

	// GetStringSlice returns the value associated with the key as a slice of strings.
	GetStringSlice(key string) []string

	// UnmarshalKey takes a single key and unmarshals it into a Struct.
	UnmarshalKey(key string, rawVal interface{}) error

	// Unmarshal unmarshals the config into a Struct. Make sure that the tags
	// on the fields of the structure are properly set.
	Unmarshal(rawVal interface{}) error

	// IsSet checks to see if the key has been set in any of the data locations.
	// IsSet is case-insensitive for a key.
	IsSet(key string) bool

	// InConfig checks to see if the given key (or an alias) is in the config file.
	InConfig(key string) bool
}

type moduleConfig struct {
	viper            *viper.Viper
	variablesConfigs map[string]VarReader // key templateFile which is used to prefix variable like ${templateFile:variableName}
}

// NewModuleConfig returns an initialized *moduleConfig.
func NewModuleConfig(core *viper.Viper, variablesConfigs map[string]VarReader) *moduleConfig {
	r := moduleConfig{
		viper:            core,
		variablesConfigs: variablesConfigs,
	}

	return &r
}

// SubReader returns new ModuleConfig instance representing a sub tree of this instance.
// SubReader is case-insensitive for a key.
func (r *moduleConfig) SubConfig(key string) ModuleConfig {
	pv := r.viper.Sub(key)

	return NewModuleConfig(pv, r.variablesConfigs)
}

func (r *moduleConfig) getString(key string) (string, bool) {
	value := r.viper.GetString(key)
	return r.replaceValue(value)
}

func (r *moduleConfig) replaceValue(value string) (string, bool) {
	var isReplaced bool
	matchStrings := rex.FindAllStringSubmatch(value, -1)
	for _, matchString := range matchStrings {
		if len(matchString) != 3 {
			continue
		}
		pattern, templateFile, variableName := matchString[0], matchString[1], matchString[2]
		variableViper, find := r.variablesConfigs[templateFile]
		if !find {
			continue
		}
		if !variableViper.IsSet(variableName) {
			continue
		}
		secretValue := variableViper.GetString(variableName)
		value = strings.Replace(value, pattern, secretValue, -1)
		isReplaced = true
	}

	return value, isReplaced
}

// GetString returns the value associated with the key as a string
func (r *moduleConfig) GetString(key string) string {
	replacedValue, _ := r.getString(key)
	return replacedValue
}

// GetBool returns the value associated with the key as a boolean.
func (r *moduleConfig) GetBool(key string) bool {
	replacedValue, ok := r.getString(key)
	if !ok {
		return r.viper.GetBool(key)
	}

	return cast.ToBool(replacedValue)
}

// GetInt64 returns the value associated with the key as an integer.
func (r *moduleConfig) GetInt64(key string) int64 {
	replacedValue, ok := r.getString(key)
	if !ok {
		return r.viper.GetInt64(key)
	}

	return cast.ToInt64(replacedValue)
}

// GetInt returns the value associated with the key as an integer.
func (r *moduleConfig) GetInt(key string) int {
	replacedValue, ok := r.getString(key)
	if !ok {
		return r.viper.GetInt(key)
	}

	return cast.ToInt(replacedValue)
}

// GetFloat64 returns the value associated with the key as a float64.
func (r *moduleConfig) GetFloat64(key string) float64 {
	replacedValue, ok := r.getString(key)
	if !ok {
		return r.viper.GetFloat64(key)
	}

	return cast.ToFloat64(replacedValue)
}

// GetTime returns the value associated with the key as time.
func (r *moduleConfig) GetTime(key string) time.Time {
	replacedValue, ok := r.getString(key)
	if !ok {
		return r.viper.GetTime(key)
	}

	return cast.ToTime(replacedValue)
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (r *moduleConfig) GetStringMapString(key string) map[string]string {
	rawMap := r.viper.GetStringMapString(key)
	for k, v := range rawMap {
		replacedValue, ok := r.replaceValue(v)
		if ok {
			rawMap[k] = replacedValue
		}
	}

	return rawMap
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (r *moduleConfig) GetStringSlice(key string) []string {
	stringSlice := r.viper.GetStringSlice(key)
	if stringSlice == nil {
		return nil
	}

	replacedSlice := make([]string, len(stringSlice))
	for i, s := range stringSlice {
		r, ok := r.replaceValue(s)
		if ok {
			replacedSlice[i] = r
		} else {
			replacedSlice[i] = s
		}
	}

	return replacedSlice
}

// UnmarshalKey takes a single key and unmarshals it into a Struct.
func (r *moduleConfig) UnmarshalKey(key string, rawVal interface{}) error {
	decodeHook := func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() == reflect.String {
			stringData := data.(string)
			replacedValue, replaced := r.replaceValue(stringData)
			if replaced {
				return replacedValue, nil
			}
		}
		return data, nil
	}

	return r.viper.UnmarshalKey(key, rawVal, viper.DecodeHook(decodeHook))
}

// Unmarshal unmarshals the config into a Struct. Make sure that the tags
// on the fields of the structure are properly set.
func (r *moduleConfig) Unmarshal(rawVal interface{}) error {
	decodeHook := func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() == reflect.String {
			stringData := data.(string)
			replacedValue, replaced := r.replaceValue(stringData)
			if replaced {
				return replacedValue, nil
			}
		}
		return data, nil
	}

	return r.viper.Unmarshal(rawVal, viper.DecodeHook(decodeHook))
}

// IsSet checks to see if the key has been set in any of the data locations.
// IsSet is case-insensitive for a key.
func (r *moduleConfig) IsSet(key string) bool {
	return r.viper.IsSet(key)
}

// InConfig checks to see if the given key (or an alias) is in the config file.
func (r *moduleConfig) InConfig(key string) bool {
	return r.viper.InConfig(key)
}

func (r *moduleConfig) GetRealViper() *viper.Viper {
	return r.viper
}
