package action

import (
	"errors"
	"net/http"
	"testing"
)

type recordingHandler struct {
	events *[]string
	result ExecutionResult
}

func (h recordingHandler) ProcessAction(*ProcessData) ExecutionResult {
	*h.events = append(*h.events, "handler")

	return h.result
}

type recordingModifier struct {
	name   string
	events *[]string
	before ExecutionResult
	after  ExecutionResult
}

func (m recordingModifier) Before(data *ProcessData) (*ProcessData, ExecutionResult) {
	*m.events = append(*m.events, "before "+m.name)

	return data, m.before
}

func (m recordingModifier) After(data *ProcessData) (*ProcessData, ExecutionResult) {
	*m.events = append(*m.events, "after "+m.name)

	return data, m.after
}

func TestModifiableHandlerRunsModifiersInWrappingOrder(t *testing.T) {
	t.Parallel()

	// Given: two successful modifiers around a successful handler.
	events := []string{}
	handler := WrapHandler(recordingHandler{events: &events}).
		WithModifier(recordingModifier{name: "first", events: &events}).
		WithModifier(recordingModifier{name: "second", events: &events})

	// When: the wrapped handler processes an action.
	result := handler.ProcessAction(&ProcessData{})

	// Then: Before unwinds and After rewinds the modifier stack.
	if result != (ExecutionResult{}) {
		t.Fatalf("unexpected result: %+v", result)
	}

	want := []string{"before second", "before first", "handler", "after first", "after second"}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}

	for index := range want {
		if events[index] != want[index] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestModifiableHandlerShortCircuitsBeforeHandler(t *testing.T) {
	t.Parallel()

	// Given: the last-added modifier returns an error from Before.
	events := []string{}
	stop := errors.New("stop")
	handler := WrapHandler(recordingHandler{events: &events}).
		WithModifier(recordingModifier{name: "first", events: &events}).
		WithModifier(recordingModifier{name: "stop", events: &events, before: ExecutionResult{Err: stop}})

	// When: the wrapper processes an action.
	result := handler.ProcessAction(&ProcessData{})

	// Then: it returns the modifier result without invoking the handler or After hooks.
	if !errors.Is(result.Err, stop) {
		t.Fatalf("result error = %v, want %v", result.Err, stop)
	}

	want := []string{"before stop"}
	if len(events) != len(want) || events[0] != want[0] {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestStaticModifiersSetStatusHeaderAndReplacePlaceholder(t *testing.T) {
	t.Parallel()

	// Given: static status/header modifiers and a post-processing placeholder modifier.
	status := NewStatusModifier().WithStatus(http.StatusAccepted)
	header := NewResponseHeaderModifier().WithHeader(http.Header{"X-Release": {"build-{version}"}})
	placeholder := NewPlaceholderModifier().WithPlaceholder("{version}").WithValue("2026.07")

	// When: they run through a Fiber request lifecycle.
	response := executeAction(t, "/", func(data *ProcessData) ExecutionResult {
		data, result := status.Before(data)
		if result.Err != nil {
			return result
		}

		data, result = header.Before(data)
		if result.Err != nil {
			return result
		}

		_, result = placeholder.After(data)

		return result
	})

	// Then: status and response header are observable by the client.
	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusAccepted)
	}

	if got := response.Header.Get("X-Release"); got != "build-2026.07" {
		t.Fatalf("X-Release = %q, want %q", got, "build-2026.07")
	}

	t.Cleanup(func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("close response body: %v", err)
		}
	})
}
