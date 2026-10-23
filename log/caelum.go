package log

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sn0wo2/CatSync/internal/util"
	"github.com/Sn0wo2/caelum"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLog(dir, logLevel, logFileFormat string) (*caelum.Logger, *lumberjack.Logger) {
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
		Writer:     util.NonClosingWriter{Writer: os.Stdout},
		Format:     caelum.Compact,
		ShowSource: true,
		TimeFormat: "2006-01-02 15:04:05",
	}}

	var file *lumberjack.Logger

	if dir != "" {
		logDir, err := filepath.Abs(dir)
		if err != nil {
			panic("failed to resolve log dir path: " + err.Error())
		}

		if err := os.MkdirAll(logDir, 0o750); err != nil {
			panic("failed to create log directory: " + err.Error())
		}

		file = &lumberjack.Logger{
			Filename:   filepath.Join(logDir, time.Now().Format(logFileFormat)),
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		}
		targets = append(targets, caelum.Target{Writer: file, Format: caelum.JSON, ShowSource: true})
	}

	return caelum.New(caelum.Config{
		Level:   level,
		Targets: targets,
	}), file
}
