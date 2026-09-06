package util

import (
	"fmt"
	"regexp"
)

func GetCompiledRegexp(pattern string) (*regexp.Regexp, error) {
	re, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return nil, fmt.Errorf("invalid regexp %q: %w", pattern, err)
	}

	return re, nil
}
