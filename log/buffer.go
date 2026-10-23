package log

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"
)

var buf struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func Writef(format string, args ...any) {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	fmt.Fprintf(&buf.buf, format, args...)
	buf.buf.WriteString("\n")
}

func Flush(logger *slog.Logger) {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	for line := range bytes.SplitSeq(buf.buf.Bytes(), []byte{'\n'}) {
		if len(line) > 0 {
			logger.Info(string(line))
		}
	}

	buf.buf.Reset()
}
