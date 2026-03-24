package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
)

func ParseLogLevel(levelStr string) (level slog.Level, err error) {
	switch strings.ToLower(strings.Trim(levelStr, "")) {
	case "debug":
		level = slog.LevelDebug
		return

	case "info":
		level = slog.LevelInfo
		return

	case "warn":
		level = slog.LevelWarn
		return

	case "error":
		level = slog.LevelError
		return
	}

	err = fmt.Errorf("invalid log levl: '%s'", levelStr)
	return
}

var logFile *os.File

func Setup(
	dispLogLevelStr string,
	fileLogLevelStr string,
) error {

	t := time.Now().Format("20060102_150405")
	var err error
	logFile, err = os.OpenFile(
		fmt.Sprintf("immich-importer.%s.log", t),
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0644,
	)
	if err != nil {
		slog.Error(
			"unable to open log file to write.",
			slog.String("error", err.Error()),
		)
		return err
	}

	defer logFile.Close()

	fileLevel, err := ParseLogLevel(fileLogLevelStr)
	if err != nil {
		err = fmt.Errorf("unable to parse log level for log file: %w", err)
		return err
	}

	dispLevel, err := ParseLogLevel(dispLogLevelStr)
	if err != nil {
		err = fmt.Errorf("unable to parse log level for log file: %w", err)
		return err
	}

	tintHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      dispLevel,
		TimeFormat: time.Kitchen,
	})
	jsonHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: fileLevel,
	})

	slog.SetDefault(slog.New(
		slog.NewMultiHandler(jsonHandler, tintHandler),
	))

	return nil
}

func CleanUp() {
	logFile.Close()
}
