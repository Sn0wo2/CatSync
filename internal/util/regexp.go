package util

import (
	"fmt"
	"regexp"

	lru "github.com/hashicorp/golang-lru/v2"
)

var regexpCache, _ = lru.New[string, *regexp.Regexp](1024)

func GetCompiledRegexp(pattern string) (*regexp.Regexp, error) {
	if re, ok := regexpCache.Get(pattern); ok {
		return re, nil
	}

	re, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return nil, fmt.Errorf("invalid regexp %q: %w", pattern, err)
	}

	regexpCache.Add(pattern, re)

	return re, nil
}
