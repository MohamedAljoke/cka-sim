// Package config fills a struct from `key`, `default`, `min`, `max` and `minLength` tags. A secret manager
// plugs in later as a Lookup for LoadWith.
package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

type Lookup func(key string) (string, bool)

func Load(target any) error {
	return LoadWith(os.LookupEnv, target)
}

func LoadWith(lookup Lookup, target any) error {
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer || value.IsNil() || value.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config: Load needs a pointer to a struct, got %T", target)
	}
	return errors.Join(fill(lookup, value.Elem())...)
}

func fill(lookup Lookup, value reflect.Value) []error {
	fields := value.Type()
	var problems []error
	for i := range fields.NumField() {
		definition := fields.Field(i)
		if definition.Anonymous && definition.Type.Kind() == reflect.Struct {
			problems = append(problems, fill(lookup, value.Field(i))...)
			continue
		}
		key, ok := definition.Tag.Lookup("key")
		if !ok {
			continue
		}
		if err := set(lookup, value.Field(i), definition, key); err != nil {
			problems = append(problems, err)
		}
	}
	return problems
}

func set(lookup Lookup, field reflect.Value, definition reflect.StructField, key string) error {
	raw, found := lookup(key)
	if !found || raw == "" {
		var hasDefault bool
		raw, hasDefault = definition.Tag.Lookup("default")
		if !hasDefault {
			return fmt.Errorf("%s is required", key)
		}
	}

	switch field.Kind() {
	case reflect.String:
		if err := withinLength(definition, key, raw); err != nil {
			return err
		}
		field.SetString(raw)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("%s must be true or false", key)
		}
		field.SetBool(parsed)
	case reflect.Int:
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%s must be a whole number", key)
		}
		if err := withinBounds(definition, key, parsed); err != nil {
			return err
		}
		field.SetInt(int64(parsed))
	case reflect.Int64:
		if field.Type() != reflect.TypeFor[time.Duration]() {
			return fmt.Errorf("%s: config cannot load into a %s", key, field.Kind())
		}
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return fmt.Errorf("%s must be a duration such as 30s or 5m", key)
		}
		field.SetInt(int64(parsed))
	default:
		return fmt.Errorf("%s: config cannot load into a %s", key, field.Kind())
	}
	return nil
}

// Never quote the value: it may be a secret.
func withinLength(definition reflect.StructField, key, value string) error {
	minLength, ok := definition.Tag.Lookup("minLength")
	if !ok {
		return nil
	}
	length, err := strconv.Atoi(minLength)
	if err != nil {
		return fmt.Errorf("%s has an unreadable minLength tag %q", key, minLength)
	}
	if len(value) < length {
		return fmt.Errorf("%s must be at least %d characters", key, length)
	}
	return nil
}

func withinBounds(definition reflect.StructField, key string, value int) error {
	if raw, ok := definition.Tag.Lookup("min"); ok {
		bound, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%s has an unreadable min tag %q", key, raw)
		}
		if value < bound {
			return fmt.Errorf("%s must be at least %d", key, bound)
		}
	}
	if raw, ok := definition.Tag.Lookup("max"); ok {
		bound, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%s has an unreadable max tag %q", key, raw)
		}
		if value > bound {
			return fmt.Errorf("%s must be at most %d", key, bound)
		}
	}
	return nil
}
