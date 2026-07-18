package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sn0wo2/CatSync/config/reader"
)

func TestApplyDefaultsFillsEmptyFieldsAndPreservesConfiguredFields(t *testing.T) {
	t.Parallel()

	defaults := GetDefaultConfig()
	configuredLog := Log{FileFormat: reader.Str("custom.log")}
	configuredServer := Server{Address: reader.Str(":8080")}
	configuredActions := []Action{}
	config := &Config{Log: configuredLog, Server: configuredServer, Actions: configuredActions}

	config = ApplyDefaults(config, defaults)

	if config.Log != configuredLog {
		t.Error("ApplyDefaults replaced a configured log")
	}

	if config.Server != configuredServer {
		t.Error("ApplyDefaults replaced a configured server")
	}

	if len(config.Actions) != 0 || config.Actions == nil {
		t.Error("ApplyDefaults replaced a configured empty actions slice")
	}

	config = ApplyDefaults(&Config{}, defaults)

	if config.Log != defaults.Log || config.Server != defaults.Server || len(config.Actions) != len(defaults.Actions) {
		t.Error("ApplyDefaults did not fill empty config fields")
	}
}

func TestDefaultConfigValidates(t *testing.T) {
	t.Parallel()

	if err := GetDefaultConfig().Validate(); err != nil {
		t.Fatalf("default config should validate: %v", err)
	}
}

func TestValidateRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*Config)
		want      string
	}{
		{
			name: "mutually exclusive ACME challenges",
			configure: func(config *Config) {
				config.Server.ACME = &ServerACME{
					Enable: true,
					Hosts:  []string{"example.com"},
					HTTP01: &ServerACMEHTTP01{},
					DNS01:  &ServerACMEDNS01{},
				}
			},
			want: "server.acme.http01 and server.acme.dns01 are mutually exclusive",
		},
		{
			name: "action without matching payload",
			configure: func(config *Config) {
				config.Actions[0].ActionServer = nil
			},
			want: "actions[0] type=server but payload is nil",
		},
		{
			name: "invalid action route regexp",
			configure: func(config *Config) {
				config.Actions[0].Route = reader.Str("[")
			},
			want: "invalid action route regexp at actions[0].route",
		},
		{
			name: "invalid action status",
			configure: func(config *Config) {
				config.Actions[3].Status = 600
			},
			want: "invalid status code at actions[3].actionModifierStatus: 600",
		},
		{
			name: "jump fallback beyond actions",
			configure: func(config *Config) {
				config.Actions[8].Fallback.JumpTo = len(config.Actions)
			},
			want: "auth fallback jumpTo out of range at actions[8]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			config := GetDefaultConfig()
			test.configure(config)

			err := config.Validate()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestActionPayloadAndModifierVisitorRespectTypeAndCallbacks(t *testing.T) {
	t.Parallel()

	action := Action{Type: ActionFile, ActionString: &ActionStringData{Content: reader.Str("wrong payload")}}
	if payload := action.GetPayload(); payload != nil {
		t.Fatalf("GetPayload() = %T, want nil for mismatched action payload", payload)
	}

	statusCalled := false
	headerCalled := false
	modifier := GlobalModifier{
		ActionModifierStatus:         &ActionModifierStatus{Status: 204},
		ActionModifierResponseHeader: &ActionModifierResponseHeader{},
	}
	modifier.EachModifier(func(m any) {
		switch mod := m.(type) {
		case *ActionModifierStatus:
			statusCalled = mod.Status == 204
		case *ActionModifierResponseHeader:
			headerCalled = true
		}
	})

	if !statusCalled || !headerCalled {
		t.Errorf("EachModifier callbacks called: status=%t header=%t, want both true", statusCalled, headerCalled)
	}
}

func TestLoaderIndexAndConfiguredPathFallback(t *testing.T) {
	if _, err := buildLoaderIndex(nil); err == nil {
		t.Fatal("buildLoaderIndex(nil) succeeded, want error")
	}

	index, err := buildLoaderIndex([]Loader{testLoader{extensions: []string{"YAML"}}})
	if err != nil {
		t.Fatalf("buildLoaderIndex() error = %v", err)
	}

	if _, ok := index[".yaml"]; !ok {
		t.Fatal("buildLoaderIndex() did not normalize an extension")
	}

	directory := t.TempDir()

	path := filepath.Join(directory, "config.yaml")
	if err := os.WriteFile(path, []byte("config"), 0o600); err != nil {
		t.Fatalf("write fallback config: %v", err)
	}

	SetConfigPath(filepath.Join(directory, "config.json"))
	t.Cleanup(func() { SetConfigPath("") })

	location := findConfigPath(index)
	if !location.Found || location.Path != path {
		t.Fatalf("findConfigPath() = %+v, want found YAML fallback at %q", location, path)
	}
}

type testLoader struct {
	extensions []string
}

func (loader testLoader) GetTag() string {
	return "test"
}

func (loader testLoader) Load(*Config, string) error {
	return errors.New("not implemented")
}

func (loader testLoader) Save(*Config, string) error {
	return errors.New("not implemented")
}

func (loader testLoader) GetAllowFileExtensions() []string {
	return loader.extensions
}
