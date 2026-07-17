//go:build catsync_all || feature_config_yaml

package loader

import (
	"path/filepath"
	"testing"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
)

func TestYAMLLoaderRoundTripsConfig(t *testing.T) {
	want := config.GetDefaultConfig()
	want.Actions[0].ActionServer.Directory = &reader.String{
		Type:    reader.StringTypePath,
		Content: "server/custom-welcome",
	}
	path := filepath.Join(t.TempDir(), "nested", "config.yaml")
	loader := NewYAMLLoader()

	if err := loader.Save(want, path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got := &config.Config{}
	if err := loader.Load(got, path); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("round-tripped config should validate: %v", err)
	}
	if len(got.Actions) != len(want.Actions) {
		t.Fatalf("round-tripped actions = %d, want %d", len(got.Actions), len(want.Actions))
	}
	if got.Server.Address.Type != reader.StringTypeString || got.Server.Address.Content != ":3000" {
		t.Fatalf("round-tripped address = %+v, want literal :3000", got.Server.Address)
	}
	directory := got.Actions[0].ActionServer.Directory
	if directory.Type != reader.StringTypePath || directory.Content != "server/custom-welcome" {
		t.Fatalf("round-tripped directory = %+v, want path-backed custom directory", directory)
	}
}
