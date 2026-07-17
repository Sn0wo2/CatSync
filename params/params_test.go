package params

import (
	"testing"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/framework"
	"go.uber.org/zap"
)

type configSourceStub struct {
	config *config.Config
}

func (s configSourceStub) CurrentConfig() *config.Config {
	return s.config
}

func TestNewHasNilGetters(t *testing.T) {
	t.Parallel()

	ctx := New()

	if ctx.GetConfig() != nil {
		t.Fatal("GetConfig() = non-nil, want nil")
	}

	if ctx.GetLogger() != nil {
		t.Fatal("GetLogger() = non-nil, want nil")
	}

	if ctx.GetFramework() != nil {
		t.Fatal("GetFramework() = non-nil, want nil")
	}
}

func TestSettersAreFluentAndExposeAssignedValues(t *testing.T) {
	t.Parallel()

	ctx := New()
	configValue := &config.Config{}
	logger := zap.NewNop()
	frameworkValue := &framework.Framework{}

	if got := ctx.SetConfigSource(configSourceStub{config: configValue}); got != ctx {
		t.Fatal("SetConfigSource() did not return the receiver")
	}

	if got := ctx.SetLogger(logger); got != ctx {
		t.Fatal("SetLogger() did not return the receiver")
	}

	if got := ctx.SetFramework(frameworkValue); got != ctx {
		t.Fatal("SetFramework() did not return the receiver")
	}

	if got := ctx.GetConfig(); got != configValue {
		t.Fatalf("GetConfig() = %p, want %p", got, configValue)
	}

	if got := ctx.GetLogger(); got != logger {
		t.Fatalf("GetLogger() = %p, want %p", got, logger)
	}

	if got := ctx.GetFramework(); got != frameworkValue {
		t.Fatalf("GetFramework() = %p, want %p", got, frameworkValue)
	}
}

func TestNilAssignmentsLeaveGettersNil(t *testing.T) {
	t.Parallel()

	ctx := New().SetConfigSource(nil).SetLogger(nil).SetFramework(nil)

	if ctx.GetConfig() != nil {
		t.Fatal("GetConfig() = non-nil after nil assignment")
	}

	if ctx.GetLogger() != nil {
		t.Fatal("GetLogger() = non-nil after nil assignment")
	}

	if ctx.GetFramework() != nil {
		t.Fatal("GetFramework() = non-nil after nil assignment")
	}
}
