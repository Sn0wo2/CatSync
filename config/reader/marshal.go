package reader

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

func (r *String) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		r.Type = StringTypeString
		r.Content = s

		return nil
	}

	type alias String

	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	r.Type = a.Type

	r.Content = a.Content
	if r.Type == "" {
		r.Type = StringTypeString
	}

	return nil
}

func (r *String) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}

	if r.Type == "" || r.Type == StringTypeString {
		return json.Marshal(r.Content)
	}

	type alias String

	return json.Marshal((*alias)(r))
}

func (r *String) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}

	if node.Kind == yaml.ScalarNode {
		r.Type = StringTypeString
		r.Content = node.Value

		return nil
	}

	type alias String

	var a alias
	if err := node.Decode(&a); err != nil {
		return err
	}

	r.Type = a.Type

	r.Content = a.Content
	if r.Type == "" {
		r.Type = StringTypeString
	}

	return nil
}

//nolint:nilnil // MarshalYAML returns nil for a nil receiver, which is the canonical YAML representation of null.
func (r *String) MarshalYAML() (any, error) {
	if r == nil {
		return nil, nil
	}

	if r.Type == "" || r.Type == StringTypeString {
		return r.Content, nil
	}

	type alias String

	return (*alias)(r), nil
}
