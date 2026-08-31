package log

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sn0wo2/caelum"
)

type nonClosingWriter struct {
	io.Writer
}

func (w nonClosingWriter) Unwrap() io.Writer {
	return w.Writer
}

func NewLog(dir, logLevel, logFileFormat string) *caelum.Logger {
	level := caelum.LevelDebug

	switch strings.ToLower(strings.TrimSpace(logLevel)) {
	case "trace":
		level = caelum.LevelTrace
	case "info":
		level = caelum.LevelInfo
	case "warn", "warning":
		level = caelum.LevelWarn
	case "error", "dpanic", "panic", "fatal":
		level = caelum.LevelError
	}

	targets := []caelum.Target{{
		Writer:     nonClosingWriter{Writer: os.Stdout},
		Format:     caelum.Compact,
		ShowSource: true,
		TimeFormat: "2006-01-02 15:04:05",
	}}

	if dir != "" {
		logDir, err := filepath.Abs(dir)
		if err != nil {
			panic("failed to resolve log dir path: " + err.Error())
		}

		if err := os.MkdirAll(logDir, 0o750); err != nil {
			panic("failed to create log directory: " + err.Error())
		}

		fileTarget := caelum.FileTarget(caelum.FileOptions{
			Path:       filepath.Join(logDir, time.Now().Format(logFileFormat)),
			MaxSizeMB:  10,
			MaxBackups: 5,
			MaxAgeDays: 30,
			Compress:   true,
		})
		fileTarget.ShowSource = true
		targets = append(targets, fileTarget)
	}

	return caelum.New(caelum.Config{
		Level:   level,
		Targets: targets,
	})
}
