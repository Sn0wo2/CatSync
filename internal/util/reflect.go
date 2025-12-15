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
		case reflect.Complex64, reflect.Complex128:
			isEmpty = value.Complex() == 0+0i
		case reflect.Ptr, reflect.Interface:
			isEmpty = value.IsNil()
		case reflect.Slice, reflect.Array:
			isEmpty = value.Len() == 0
			if !isEmpty {
				if value.Len() > 0 {
					elem := value.Index(0)

					elemKind := elem.Kind()
					if elemKind == reflect.Struct || (elemKind == reflect.Ptr && elem.Elem().Kind() == reflect.Struct) {
						slicePath := field.Name
						if path != "" {
							slicePath = path + "." + slicePath
						}

						sliceTagPath := getFieldNameForPath(field, tag)
						if tagPath != "" {
							sliceTagPath = tagPath + "." + sliceTagPath
						}

						for i := range value.Len() {
							elementPath := fmt.Sprintf("%s[%d]", slicePath, i)

							elementTagPath := fmt.Sprintf("%s[%d]", sliceTagPath, i)
							if err := validateWithPath(value.Index(i).Interface(), tag, elementPath, elementTagPath); err != nil {
								return err
							}
						}
					}
				}
			}
		case reflect.Map:
			isEmpty = value.Len() == 0
			if !isEmpty {
				mapPath := field.Name
				if path != "" {
					mapPath = path + "." + mapPath
				}

				mapTagPath := getFieldNameForPath(field, tag)
				if tagPath != "" {
					mapTagPath = tagPath + "." + mapTagPath
				}

				iter := value.MapRange()
				for iter.Next() {
					k := iter.Key()
					v := iter.Value()

					if v.IsValid() && (v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && !v.IsNil() && v.Elem().Kind() == reflect.Struct)) {
						elementPath := fmt.Sprintf("%s[%v]", mapPath, k.Interface())

						elementTagPath := fmt.Sprintf("%s[%v]", mapTagPath, k.Interface())
						if err := validateWithPath(v.Interface(), tag, elementPath, elementTagPath); err != nil {
							return err
						}
					}
				}
			}
		case reflect.Struct:
			if value.CanAddr() {
				newPath := field.Name
				if path != "" {
					newPath = path + "." + newPath
				}

				newTagPath := getFieldNameForPath(field, tag)
				if tagPath != "" {
					newTagPath = tagPath + "." + newTagPath
				}

				if err := validateWithPath(value.Addr().Interface(), tag, newPath, newTagPath); err != nil {
					return err
				}
			}
		case reflect.Chan, reflect.Func, reflect.UnsafePointer:
			isEmpty = value.IsNil()
		default:
		}

		if !isEmpty {
			continue
		}

		fullPath := field.Name
		if path != "" {
			fullPath = path + "." + fullPath
		}

		fullTagPath := getFieldNameForPath(field, tag)
		if tagPath != "" {
			fullTagPath = tagPath + "." + fullTagPath
		}

		errMsg := fmt.Sprintf("validate: field [%s] is required but empty", fullPath)
		if fullTagPath != "" {
			errMsg = fmt.Sprintf("validate: field [%s](%s) is required but empty", fullPath, fullTagPath)
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
