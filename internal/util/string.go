package util

func TitleCase(s string) string {
	if len(s) == 0 {
		return s
	}

	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}

	return s
}
