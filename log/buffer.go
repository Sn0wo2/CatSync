package log

import (
	"bytes"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type Buffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

var buf Buffer

func Write(format string, args ...interface{}) {
	buf.mu.Lock()
	defer buf.mu.Unlock()
	buf.buf.WriteString(fmt.Sprintf(format, args...))
	buf.buf.WriteString("\n")
}

func Flush(logger *zap.Logger) {
	buf.mu.Lock()
	defer buf.mu.Unlock()
	if buf.buf.Len() > 0 {
		for _, line := range bytes.Split(buf.buf.Bytes(), []byte{'\n'}) {
			if len(line) > 0 {
				logger.Info(string(line))
			}
		}
		buf.buf.Reset()
	}
}
