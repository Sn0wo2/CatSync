package util

import (
	"errors"
	"regexp"
	"strings"
	"sync"
	"testing"
)

func TestGetCompiledRegexpAnchorsPattern(t *testing.T) {
	t.Parallel()

	re, err := GetCompiledRegexp("cat")
	if err != nil {
		t.Fatalf("GetCompiledRegexp() error = %v", err)
	}

	if !re.MatchString("cat") {
		t.Fatal("compiled regexp did not match the full pattern")
	}

	if re.MatchString("cats") {
		t.Fatal("compiled regexp matched a suffixed string")
	}

	if re.MatchString("copycat") {
		t.Fatal("compiled regexp matched a prefixed string")
	}
}

func TestGetCompiledRegexpReturnsWrappedCompileError(t *testing.T) {
	t.Parallel()

	re, err := GetCompiledRegexp("[")
	if re != nil {
		t.Fatalf("GetCompiledRegexp() regexp = %v, want nil", re)
	}

	if err == nil {
		t.Fatal("GetCompiledRegexp() error = nil, want compile error")
	}

	if !strings.Contains(err.Error(), `invalid regexp "[":`) {
		t.Fatalf("GetCompiledRegexp() error = %q, want contextual pattern", err)
	}

	if errors.Unwrap(err) == nil {
		t.Fatal("GetCompiledRegexp() error does not wrap the compile error")
	}
}

func TestGetCompiledRegexpCachesPointerByPattern(t *testing.T) {
	t.Parallel()

	first, err := GetCompiledRegexp("cache-identity")
	if err != nil {
		t.Fatalf("first GetCompiledRegexp() error = %v", err)
	}

	second, err := GetCompiledRegexp("cache-identity")
	if err != nil {
		t.Fatalf("second GetCompiledRegexp() error = %v", err)
	}

	if first != second {
		t.Fatalf("cache returned different regexp pointers: %p and %p", first, second)
	}
}

func TestGetCompiledRegexpIsSafeForConcurrentCalls(t *testing.T) {
	t.Parallel()

	const callers = 64

	regexps := make(chan *regexp.Regexp, callers)
	errs := make(chan error, callers)

	var waitGroup sync.WaitGroup

	for range callers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			re, err := GetCompiledRegexp("concurrent-cache")
			if err == nil && !re.MatchString("concurrent-cache") {
				errs <- errors.New("compiled regexp did not match its pattern")

				return
			}

			if err != nil {
				errs <- err

				return
			}

			regexps <- re
		}()
	}

	waitGroup.Wait()
	close(regexps)
	close(errs)

	for err := range errs {
		t.Fatalf("concurrent GetCompiledRegexp() error = %v", err)
	}

	var first *regexp.Regexp
	for re := range regexps {
		if first == nil {
			first = re

			continue
		}

		if re != first {
			t.Fatalf("concurrent cache returned different regexp pointers: %p and %p", re, first)
		}
	}
}
