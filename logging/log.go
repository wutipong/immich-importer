package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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
var currentLogFilePath string

func Setup(
	profile string,
	dispLogLevelStr string,
	enableFileLog bool,
	fileLogLevelStr string,
) error {
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
		// AddSource:  true,
	})

	logLocation, err := CreateLogDirectoryPath()
	if err != nil {
		return fmt.Errorf("unable to create log file path: %w", err)
	}

	err = os.MkdirAll(logLocation, 0755)
	if err != nil {
		return fmt.Errorf("unable to create log directory: %w", err)
	}

	if !enableFileLog {
		slog.SetDefault(slog.New(tintHandler))
		return nil
	} else {
		currentLogFilePath = filepath.Join(logLocation, CreateLogFileName(profile))
		logFile, err = os.OpenFile(
			currentLogFilePath,
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

		jsonHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level: fileLevel,
			// AddSource: true,
		})

		slog.SetDefault(slog.New(
			slog.NewMultiHandler(jsonHandler, tintHandler),
		))
	}

	return nil
}

func CleanUp() error {
	if logFile == nil {
		return nil
	}

	err := logFile.Close()
	if err != nil {
		return fmt.Errorf("failed to clean up logging: %w", err)
	}

	return nil
}

func CreateLogDirectoryPath() (path string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error(
			"failed to get user home directory",
			slog.String("error", err.Error()),
		)
		return
	}
	path = filepath.Join(homeDir, ".immich-importer", "logs")

	return
}

func CreateLogFileName(profile string) string {
	t := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s.log", profile, t)
}
