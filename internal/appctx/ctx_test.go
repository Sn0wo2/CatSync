package appctx

import (
	"testing"

	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/framework"
	"go.uber.org/zap"
)

type configSourceStub struct {
	cfg *config.Config
}

func (s configSourceStub) CurrentConfig() *config.Config {
	return s.cfg
}

func TestNewHasNilGetters(t *testing.T) {
	t.Parallel()

	ctx := New()

	if ctx.GetConfig() != nil {
		t.Fatal("GetConfig() = non-nil, want nil")
	}

	if ctx.Logger != nil {
		t.Fatal("Logger = non-nil, want nil")
	}

	if ctx.FW != nil {
		t.Fatal("FW = non-nil, want nil")
	}
}

func TestSettersAreFluentAndExposeAssignedValues(t *testing.T) {
	t.Parallel()

	ctx := New()
	configValue := &config.Config{}
	logger := zap.NewNop()
	frameworkValue := &framework.Framework{}

	ctx.ConfigSource = configSourceStub{cfg: configValue}
	ctx.Logger = logger
	ctx.FW = frameworkValue

	if got := ctx.GetConfig(); got != configValue {
		t.Fatalf("GetConfig() = %p, want %p", got, configValue)
	}

	if got := ctx.Logger; got != logger {
		t.Fatalf("Logger = %p, want %p", got, logger)
	}

	if got := ctx.FW; got != frameworkValue {
		t.Fatalf("FW = %p, want %p", got, frameworkValue)
	}
}

func TestNilAssignmentsLeaveGettersNil(t *testing.T) {
	t.Parallel()

	ctx := New()
	ctx.ConfigSource = nil
	ctx.Logger = nil
	ctx.FW = nil

	if ctx.GetConfig() != nil {
		t.Fatal("GetConfig() = non-nil after nil assignment")
	}

	if ctx.Logger != nil {
		t.Fatal("Logger = non-nil after nil assignment")
	}

	if ctx.FW != nil {
		t.Fatal("FW = non-nil after nil assignment")
	}
}
