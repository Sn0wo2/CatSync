package util

import "io"

type NonClosingWriter struct{ io.Writer }

func (w NonClosingWriter) Unwrap() io.Writer {
	return w.Writer
}
