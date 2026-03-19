package util

import (
	"fmt"
	"regexp"
	"sync"
)

var (
	regexpCache = make(map[string]*regexp.Regexp)
	cacheMutex  sync.RWMutex
)

func GetCompiledRegexp(pattern string) (*regexp.Regexp, error) {
	cacheMutex.RLock()

	if re, ok := regexpCache[pattern]; ok {
		cacheMutex.RUnlock()

		return re, nil
	}

	cacheMutex.RUnlock()

	re, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return nil, fmt.Errorf("invalid regexp %q: %w", pattern, err)
	}

	cacheMutex.Lock()

	regexpCache[pattern] = re

	cacheMutex.Unlock()

	return re, nil
}
