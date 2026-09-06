//go:build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
	"github.com/invopop/jsonschema"
	orderedmap "github.com/pb33f/ordered-map/v2"
)

func main() {
	output := flag.String("o", "config.schema.json", "output schema file path")
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	absOutput := filepath.Join(wd, *output)

	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		Mapper: func(t reflect.Type) *jsonschema.Schema {
			if t != reflect.TypeOf(reader.String{}) {
				return nil
			}

			properties := orderedmap.New[string, *jsonschema.Schema]()
			properties.Set("type", &jsonschema.Schema{
				Type: "string",
				Enum: []any{
					reader.StringTypeAuto,
					reader.StringTypeString,
					reader.StringTypePath,
					reader.StringTypeHTTP,
				},
			})
			properties.Set("content", &jsonschema.Schema{Type: "string"})

			return &jsonschema.Schema{
				OneOf: []*jsonschema.Schema{
					{Type: "string"},
					{
						Type:                 "object",
						Properties:           properties,
						AdditionalProperties: jsonschema.FalseSchema,
						Required:             []string{"content"},
					},
					{Type: "null"},
				},
			}
		},
	}

	schema := reflector.Reflect(&config.Config{})
	schema.Title = "CatSync Config"
	schema.Description = "Configuration schema for CatSync HTTP server"

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling schema: %v\n", err)
		os.Exit(1)
	}

	dir := filepath.Dir(absOutput)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	f, err := os.Create(absOutput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s\n", absOutput)
}
