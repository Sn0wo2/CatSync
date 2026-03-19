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

func Merge[T any](dst, src *T) {
	srcVal := reflect.ValueOf(src).Elem()
	srcType := srcVal.Type()
	dstVal := reflect.ValueOf(dst).Elem()

	for i := range srcVal.NumField() {
		srcField := srcVal.Field(i)
		if optional := srcType.Field(i).Tag.Get("optional"); optional == "true" {
			continue
		}

		dstField := dstVal.Field(i)
		if dstField.IsZero() && !srcField.IsZero() {
			dstField.Set(srcField)
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
	case yaml.AliasNode:
		mergeYamlNodeRecursive(node.Alias, value)
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
			case reflect.Invalid,
				reflect.Uintptr,
				reflect.Complex64,
				reflect.Complex128,
				reflect.Array,
				reflect.Chan,
				reflect.Func,
				reflect.Interface,
				reflect.Map,
				reflect.Ptr,
				reflect.Slice,
				reflect.Struct,
				reflect.UnsafePointer:
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

func getFieldNameForPath(field reflect.StructField, tag string) string {
	tagValue := field.Tag.Get(tag)
	if tagValue != "" {
		return tagValue
	}

	return field.Name
}

func Validate(cfg any, tag string) error {
	return validateWithPath(cfg, tag, "", "")
}

func validateWithPath(cfg any, tag string, path string, tagPath string) error {
	v, err := getStructValue(cfg)
	if err != nil {
		return err
	}

	return validateStructValue(v, tag, path, tagPath)
}

func getStructValue(cfg any) (reflect.Value, error) {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}, errors.New("validate: nil pointer received")
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("validate: input must be a struct or a pointer to a struct, but got %s", v.Kind())
	}

	return v, nil
}

func validateStructValue(v reflect.Value, tag string, path string, tagPath string) error {
	t := v.Type()
	for i := range v.NumField() {
		field := t.Field(i)
		if field.PkgPath != "" || field.Tag.Get("optional") == "true" {
			continue
		}

		fieldPath, fieldTagPath := buildFieldPaths(field, tag, path, tagPath)

		isEmpty, err := validateFieldValue(v.Field(i), tag, fieldPath, fieldTagPath)
		if err != nil {
			return err
		}

		if isEmpty {
			return buildRequiredFieldError(t, field, fieldPath, fieldTagPath)
		}
	}

	return nil
}

func buildFieldPaths(field reflect.StructField, tag string, path string, tagPath string) (string, string) {
	fieldPath := field.Name
	if path != "" {
		fieldPath = path + "." + fieldPath
	}

	fieldTagPath := getFieldNameForPath(field, tag)
	if tagPath != "" {
		fieldTagPath = tagPath + "." + fieldTagPath
	}

	return fieldPath, fieldTagPath
}

func validateFieldValue(value reflect.Value, tag string, path string, tagPath string) (bool, error) {
	switch value.Kind() {
	case reflect.Invalid:
		return true, nil
	case reflect.Bool:
		return !value.Bool(), nil
	case reflect.String:
		return strings.TrimSpace(value.String()) == "", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0, nil
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0, nil
	case reflect.Complex64, reflect.Complex128:
		return value.Complex() == 0+0i, nil
	case reflect.Ptr, reflect.Interface:
		return value.IsNil(), nil
	case reflect.Slice, reflect.Array:
		return validateSequenceValue(value, tag, path, tagPath)
	case reflect.Map:
		return validateMapValue(value, tag, path, tagPath)
	case reflect.Struct:
		if !value.CanAddr() {
			return false, nil
		}

		return false, validateWithPath(value.Addr().Interface(), tag, path, tagPath)
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return value.IsNil(), nil
	default:
		return false, nil
	}
}

func validateSequenceValue(value reflect.Value, tag string, path string, tagPath string) (bool, error) {
	if value.Len() == 0 {
		return true, nil
	}

	for i := range value.Len() {
		element := value.Index(i)
		if !isStructLikeValue(element) {
			continue
		}

		elementPath := fmt.Sprintf("%s[%d]", path, i)

		elementTagPath := fmt.Sprintf("%s[%d]", tagPath, i)
		if err := validateWithPath(element.Interface(), tag, elementPath, elementTagPath); err != nil {
			return false, err
		}
	}

	return false, nil
}

func validateMapValue(value reflect.Value, tag string, path string, tagPath string) (bool, error) {
	if value.Len() == 0 {
		return true, nil
	}

	iter := value.MapRange()
	for iter.Next() {
		key := iter.Key()

		element := iter.Value()
		if !isStructLikeValue(element) {
			continue
		}

		elementPath := fmt.Sprintf("%s[%v]", path, key.Interface())

		elementTagPath := fmt.Sprintf("%s[%v]", tagPath, key.Interface())
		if err := validateWithPath(element.Interface(), tag, elementPath, elementTagPath); err != nil {
			return false, err
		}
	}

	return false, nil
}

func isStructLikeValue(value reflect.Value) bool {
	if !value.IsValid() {
		return false
	}

	if value.Kind() == reflect.Struct {
		return true
	}

	return value.Kind() == reflect.Ptr && !value.IsNil() && value.Elem().Kind() == reflect.Struct
}

func buildRequiredFieldError(parentType reflect.Type, field reflect.StructField, path string, tagPath string) error {
	errMsg := fmt.Sprintf("validate: field [%s] is required but empty", path)
	if tagPath != "" {
		errMsg = fmt.Sprintf("validate: field [%s](%s) is required but empty", path, tagPath)
	}

	if defaultValue, ok := getDefaultFieldValue(parentType, field.Name); ok {
		errMsg += fmt.Sprintf(", default value is: '%v'", defaultValue)
	}

	return errors.New(errMsg)
}

func getDefaultFieldValue(parentType reflect.Type, fieldName string) (any, bool) {
	if DefaultConfigProvider == nil {
		return nil, false
	}

	defaultCfg, ok := DefaultConfigProvider()
	if !ok {
		return nil, false
	}

	defaultValue := reflect.ValueOf(defaultCfg)
	if defaultValue.Kind() != reflect.Ptr || defaultValue.IsNil() {
		return nil, false
	}

	parentStruct := defaultValue.Elem().FieldByName(parentType.Name())
	if !parentStruct.IsValid() {
		return nil, false
	}

	fieldValue := parentStruct.FieldByName(fieldName)
	if !fieldValue.IsValid() {
		return nil, false
	}

	return fieldValue.Interface(), true
}
