package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = newValidator()

func newValidator() *validator.Validate {
	val := validator.New(validator.WithRequiredStructEnabled())
	// errors name a field the way the config file does
	val.RegisterTagNameFunc(tomlName)
	return val
}

func tomlName(f reflect.StructField) string {
	name, _, _ := strings.Cut(f.Tag.Get("toml"), ",")
	if name == "-" {
		return ""
	}
	return name
}

func Validate(s any) error {
	err := v.Struct(s)
	errs, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok {
		return err
	}
	msgs := make([]string, 0, len(errs))
	for _, fe := range errs {
		msgs = append(msgs, message(reflect.TypeOf(s), fe))
	}
	return fmt.Errorf("invalid config: %s", strings.Join(msgs, "; "))
}

func message(root reflect.Type, fe validator.FieldError) string {
	// the namespace starts with the struct's own type name
	_, field, _ := strings.Cut(fe.Namespace(), ".")
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "required_with":
		return fmt.Sprintf("%s is required when %s is set", field, siblingName(root, fe))
	case "min":
		return fmt.Sprintf("%s needs at least %s", field, fe.Param())
	case "url":
		return field + " must be a URL"
	default:
		return fmt.Sprintf("%s fails %q", field, fe.Tag())
	}
}

// required_with names a field by its go name, a message wants its config name
func siblingName(root reflect.Type, fe validator.FieldError) string {
	t := root
	path := strings.Split(fe.StructNamespace(), ".")
	// skip the root type's name and the failing field itself
	for _, seg := range path[1 : len(path)-1] {
		name, _, _ := strings.Cut(seg, "[")
		f, ok := deref(t).FieldByName(name)
		if !ok {
			return fe.Param()
		}
		t = f.Type
	}
	if f, ok := deref(t).FieldByName(fe.Param()); ok {
		return tomlName(f)
	}
	return fe.Param()
}

// the struct behind pointers and the values of maps and slices
func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Map || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	return t
}
