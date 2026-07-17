package router

import (
	"testing"

	"github.com/Sn0wo2/CatSync/params"
)

func TestInit_PanicsWhenFrameworkIsNil(t *testing.T) {
	t.Parallel()

	// Given: a context without a framework.
	ctx := params.New()

	// When: the router is initialized.
	defer func() {
		// Then: initialization rejects the missing framework immediately.
		if recovered := recover(); recovered != "framework not found" {
			t.Fatalf("Init() panic = %v, want %q", recovered, "framework not found")
		}
	}()

	Init(ctx, nil)
}
