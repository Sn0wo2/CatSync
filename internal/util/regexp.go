package util

import (
	"fmt"
	"regexp"

	"github.com/Sn0wo2/CatSync/internal/cache"
)

var regexpCache, _ = cache.NewLRU[string, *regexp.Regexp](1024)

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
