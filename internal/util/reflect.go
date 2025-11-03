package util

import (
	"fmt"
	"reflect"
	"strings"
)

func Merge[T any](dst, src *T) {
	dstVal := reflect.ValueOf(dst).Elem()
	srcVal := reflect.ValueOf(src).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		if !srcField.IsZero() {
			dstVal.Field(i).Set(srcField)
		}
	}
}

func Validate(stru any) error {
	v := reflect.ValueOf(stru)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fmt.Errorf("validate: nil pointer received")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("validate: input must be a struct or a pointer to a struct, but got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // Skip unexported fields
			continue
		}

		if optional := field.Tag.Get("optional"); optional == "true" {
			continue
		}

		value := v.Field(i)

		var isEmpty bool

		//nolint:exhaustive
		switch value.Kind() {
		case reflect.String:
			isEmpty = strings.TrimSpace(value.String()) == ""
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			isEmpty = value.Int() == 0
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			isEmpty = value.Uint() == 0
		case reflect.Float32, reflect.Float64:
			isEmpty = value.Float() == 0
		case reflect.Ptr, reflect.Interface:
			isEmpty = value.IsNil()
		case reflect.Slice, reflect.Map, reflect.Array:
			isEmpty = value.Len() == 0
		case reflect.Struct:
			if value.CanAddr() {
				if err := Validate(value.Addr().Interface()); err != nil {
					return fmt.Errorf("%s.%s: %w", t.Name(), field.Name, err)
				}
			}
		default:
		}

		if isEmpty {
			return fmt.Errorf("validate: field [%s] is required but empty", field.Name)
		}
	}
	return nil
}
