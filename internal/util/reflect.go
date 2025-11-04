package util

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

func Merge[T any](dst, src *T) {
	dstVal := reflect.ValueOf(dst).Elem()
	srcVal := reflect.ValueOf(src).Elem()
	srcType := srcVal.Type()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		if optional := srcType.Field(i).Tag.Get("optional"); optional == "true" {
			continue
		}

		if !srcField.IsZero() {
			dstVal.Field(i).Set(srcField)
		}
	}
}

func MergeYamlNode(node *yaml.Node, cfg any) {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	mergeYamlNodeRecursive(node, v)
}

func mergeYamlNodeRecursive(node *yaml.Node, value reflect.Value) {
	if node == nil {
		return
	}

	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return
		}
		value = value.Elem()
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, n := range node.Content {
			mergeYamlNodeRecursive(n, value)
		}
	case yaml.MappingNode:
		if value.Kind() != reflect.Struct {
			return
		}
		t := value.Type()
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			fieldName := keyNode.Value
			var field reflect.StructField
			var found bool
			for j := 0; j < t.NumField(); j++ {
				tag := t.Field(j).Tag.Get("yaml")
				if tag == fieldName {
					field = t.Field(j)
					found = true
					break
				}
			}

			if found {
				fieldValue := value.FieldByName(field.Name)
				mergeYamlNodeRecursive(valueNode, fieldValue)
			}
		}
	case yaml.ScalarNode:
		if value.CanSet() {
			var newValue string
			currentValue := node.Value

			switch value.Kind() {
			case reflect.String:
				newValue = value.String()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				newValue = fmt.Sprintf("%d", value.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				newValue = fmt.Sprintf("%d", value.Uint())
			case reflect.Float32, reflect.Float64:
				newValue = fmt.Sprintf("%g", value.Float()) // Use %g for cleaner float output
			case reflect.Bool:
				newValue = fmt.Sprintf("%t", value.Bool())
			default:
				// For other types, don't attempt to update the node value
				return
			}

			if currentValue != newValue {
				node.Value = newValue
			}
		}
	case yaml.SequenceNode:
		if value.Kind() == reflect.Slice {
			for i := 0; i < value.Len() && i < len(node.Content); i++ {
				mergeYamlNodeRecursive(node.Content[i], value.Index(i))
			}
		}
	default:
	}
}

func Validate(stru any) error {
	v := reflect.ValueOf(stru)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return errors.New("validate: nil pointer received")
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("validate: input must be a struct or a pointer to a struct, but got %s", v.Kind())
	}

	t := v.Type()
	for i := range v.NumField() {
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
			// Do nothing for other types.
		}

		if isEmpty {
			return fmt.Errorf("validate: field [%s] is required but empty", field.Name)
		}
	}

	return nil
}
