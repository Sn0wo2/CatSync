package util

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var DefaultConfigProvider func() (any, bool)

func ReplaceVersionInAny(data any, version string) any {
	if data == nil {
		return nil
	}

	if m, ok := data.(map[string]any); ok {
		result := make(map[string]any)
		for k, v := range m {
			result[k] = ReplaceVersionInAny(v, version)
		}
		return result
	}

	if s, ok := data.([]any); ok {
		result := make([]any, len(s))
		for i, v := range s {
			result[i] = ReplaceVersionInAny(v, version)
		}
		return result
	}

	if s, ok := data.(string); ok && s == "{{version}}" {
		return version
	}

	return data
}

func Merge[T any](dst, src *T) {
	dstVal := reflect.ValueOf(dst).Elem()
	srcVal := reflect.ValueOf(src).Elem()
	srcType := srcVal.Type()

	for i := range srcVal.NumField() {
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

	//nolint:exhaustive
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

			var (
				field reflect.StructField
				found bool
			)

			for j := range t.NumField() {
				if t.Field(j).Tag.Get("yaml") == fieldName {
					field = t.Field(j)
					found = true

					break
				}
			}

			if found {
				mergeYamlNodeRecursive(valueNode, value.FieldByName(field.Name))
			}
		}
	case yaml.ScalarNode:
		if value.CanSet() {
			var newValue string
			//nolint:exhaustive
			switch value.Kind() {
			case reflect.String:
				newValue = value.String()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				newValue = strconv.FormatInt(value.Int(), 10)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				newValue = strconv.FormatUint(value.Uint(), 10)
			case reflect.Float32, reflect.Float64:
				// Use %g for cleaner float output, avoiding trailing zeros.
				newValue = fmt.Sprintf("%g", value.Float())
			case reflect.Bool:
				newValue = strconv.FormatBool(value.Bool())
			default:
				// Unsupported types are ignored to prevent incorrect mutations.
				return
			}

			if node.Value != newValue && newValue != "" {
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

func Validate(cfg any, tag string) error {
	v := reflect.ValueOf(cfg)
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

		if field.PkgPath != "" || field.Tag.Get("optional") == "true" {
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
				if err := Validate(value.Addr().Interface(), tag); err != nil {
					return err
				}
			}
		default:
			// Other types are not checked for emptiness
		}

		if !isEmpty {
			continue
		}

		tagValue := field.Tag.Get(tag)

		errMsg := fmt.Sprintf("validate: field [%s] is required but empty", field.Name)
		if tagValue != "" {
			errMsg = fmt.Sprintf("validate: field [%s](%s) is required but empty", field.Name, tagValue)
		}

		if DefaultConfigProvider != nil {
			if defaultCfg, ok := DefaultConfigProvider(); ok {
				parentStruct := reflect.ValueOf(defaultCfg).Elem().FieldByName(t.Name())
				if parentStruct.IsValid() {
					defaultValue := parentStruct.FieldByName(field.Name)
					if defaultValue.IsValid() {
						errMsg += fmt.Sprintf(", default value is: '%v'", defaultValue.Interface())
					}
				}
			}
		}

		return errors.New(errMsg)
	}

	return nil
}
