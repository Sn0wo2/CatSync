package execute

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sn0wo2/CatSync/action"
	"github.com/Sn0wo2/CatSync/config"
	"github.com/Sn0wo2/CatSync/config/reader"
	"github.com/Sn0wo2/CatSync/internal/appctx"
	"github.com/gofiber/fiber/v3"
)

type statusHandler map[string]action.ExecutionResult

const (
	testCompleteOutcome = "complete"
	testExactRoute      = "/health"
	testNextOutcome     = "next"
)

func (h statusHandler) ProcessAction(data *action.ProcessData) action.ExecutionResult {
	stringData, ok := data.Payload.(*config.ActionStringData)
	if !ok {
		return action.ExecutionResult{Err: fmt.Errorf("unexpected payload type %T", data.Payload)}
	}

	return h[stringData.Content.Must()]
}

func stringAction(route, outcome string) config.Action {
	return config.Action{
		Route: reader.Str(route),
		Type:  config.ActionString,
		ActionString: &config.ActionStringData{
			Content: reader.Str(outcome),
		},
	}
}

func buildExecutor(t *testing.T, actions []config.Action, handler action.Handler) *Executor {
	t.Helper()

	reg := action.Registry{
		config.ActionString: handler,
	}

	executor, err := NewWithRegistry(reg).WithConfig(&config.Config{Actions: actions}).Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	return executor
}

func dispatch(t *testing.T, executor *Executor, path string) (bool, error) {
	t.Helper()

	app := fiber.New()

	var (
		matched     bool
		dispatchErr error
	)

	app.Use(func(c fiber.Ctx) error {
		matched, dispatchErr = executor.Dispatch(&RequestContext{
			Ctx:      appctx.New(),
			FiberCtx: c,
		})

		return nil
	})

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "http://catsync.test"+path, nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}

	t.Cleanup(func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("close response body: %v", err)
		}
	})

	return matched, dispatchErr
}

func TestBuildRejectsInvalidActionShapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		action  config.Action
		wantErr string
	}{
		{
			name: "nonliteral route",
			action: config.Action{
				Route: &reader.String{Type: reader.StringTypePath, Content: "route.txt"},
				Type:  config.ActionString,
				ActionString: &config.ActionStringData{
					Content: reader.Str(testCompleteOutcome),
				},
			},
			wantErr: "route must be a literal string",
		},
		{
			name:    "invalid route regexp",
			action:  stringAction("[", testCompleteOutcome),
			wantErr: "invalid action route regexp",
		},
		{
			name: "unknown handler",
			action: config.Action{
				Route:        reader.Str("/known"),
				Type:         config.ActionType("unknown"),
				ActionString: &config.ActionStringData{Content: reader.Str(testCompleteOutcome)},
			},
			wantErr: "unknown action handler",
		},
		{
			name: "missing payload",
			action: config.Action{
				Route: reader.Str("/missing-payload"),
				Type:  config.ActionString,
			},
			wantErr: "payload is nil",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			reg := action.NewRegistry()

			_, err := NewWithRegistry(reg).WithConfig(&config.Config{Actions: []config.Action{test.action}}).Build()
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("Build() error = %v, want containing %q", err, test.wantErr)
			}
		})
	}
}

func TestDispatchMatchesExactAndRegularExpressionRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		route   string
		path    string
		matched bool
	}{
		{name: "exact route", route: testExactRoute, path: testExactRoute, matched: true},
		{name: "exact route is not a prefix", route: testExactRoute, path: "/healthz", matched: false},
		{name: "regular expression route", route: "/users/[0-9]+", path: "/users/42", matched: true},
		{name: "regular expression route is anchored", route: "/users/[0-9]+", path: "/users/42/profile", matched: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			executor := buildExecutor(t, []config.Action{
				stringAction(test.route, testCompleteOutcome),
				stringAction("", testNextOutcome),
			}, statusHandler{
				testCompleteOutcome: {},
				testNextOutcome:     {Status: action.ExecutionFallbackNext},
			})

			matched, err := dispatch(t, executor, test.path)
			if err != nil {
				t.Fatalf("Dispatch() error = %v", err)
			}

			if matched != test.matched {
				t.Fatalf("Dispatch() matched = %t, want %t", matched, test.matched)
			}
		})
	}
}

func TestDispatchContinuesAfterFallbackNext(t *testing.T) {
	t.Parallel()

	executor := buildExecutor(t, []config.Action{
		stringAction("/resource", testNextOutcome),
		stringAction("/resource", testCompleteOutcome),
	}, statusHandler{
		testNextOutcome:     {Status: action.ExecutionFallbackNext},
		testCompleteOutcome: {},
	})

	matched, err := dispatch(t, executor, "/resource")
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	if !matched {
		t.Fatal("Dispatch() matched = false, want true after fallback next")
	}
}

func TestDispatchRejectsInvalidAndLoopingFallbackJumps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		actions []config.Action
		handler statusHandler
		wantErr string
	}{
		{
			name:    "invalid target",
			actions: []config.Action{stringAction("/entry", "jump-out-of-range")},
			handler: statusHandler{"jump-out-of-range": {Status: action.ExecutionFallbackJump, JumpTo: 3}},
			wantErr: "invalid auth fallback jumpTo index: 3",
		},
		{
			name: "looping targets",
			actions: []config.Action{
				stringAction("/entry", "jump-one"),
				stringAction("/unreachable", "jump-zero"),
			},
			handler: statusHandler{
				"jump-one":  {Status: action.ExecutionFallbackJump, JumpTo: 1},
				"jump-zero": {Status: action.ExecutionFallbackJump, JumpTo: 0},
			},
			wantErr: "auth fallback jump loop detected at index: 1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			executor := buildExecutor(t, test.actions, test.handler)

			matched, err := dispatch(t, executor, "/entry")
			if matched {
				t.Fatal("Dispatch() matched = true, want false")
			}

			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("Dispatch() error = %v, want containing %q", err, test.wantErr)
			}
		})
	}
}

func TestDispatchExecutesFallbackJumpWithoutMatchingTargetRoute(t *testing.T) {
	t.Parallel()

	executor := buildExecutor(t, []config.Action{
		stringAction("/entry", "jump"),
		stringAction("/unreachable", testCompleteOutcome),
	}, statusHandler{
		"jump":              {Status: action.ExecutionFallbackJump, JumpTo: 1},
		testCompleteOutcome: {},
	})

	matched, err := dispatch(t, executor, "/entry")
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	if !matched {
		t.Fatal("Dispatch() matched = false, want true after fallback jump")
	}
}

func TestDispatchReturnsFalseWhenNoActionMatches(t *testing.T) {
	t.Parallel()

	executor := buildExecutor(t, []config.Action{
		stringAction("/configured", "next"),
	}, statusHandler{
		"next": {Status: action.ExecutionFallbackNext},
	})

	matched, err := dispatch(t, executor, "/other")
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}

	if matched {
		t.Fatal("Dispatch() matched = true, want false when no action matches")
	}
}
