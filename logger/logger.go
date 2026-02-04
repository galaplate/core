package logger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

type logger struct {
	mu     sync.Mutex
	file   *os.File
	level  slog.Level
	output string
}

type logMessage struct {
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Data      map[string]any `json:"additional_info,omitempty"`
}

var logInstance *logger

func init() {
	logInstance = &logger{level: slog.LevelInfo}

	logsDir := os.Getenv("GALAPLATE_LOGS_DIR")
	if logsDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get working directory: %v\n", err)
			return
		}
		logsDir = cwd + "/storage/logs"
	}

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logs directory: %v\n", err)
		return
	}

	logFilename := fmt.Sprintf("%s/app.%s.log", logsDir, time.Now().Format("2006-01-02"))

	if err := logInstance.SetFile(logFilename); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set log file: %v\n", err)
		return
	}

	_, err := rotatelogs.New(
		logsDir+"/app.%Y-%m-%d.log",
		rotatelogs.WithLinkName(logsDir+"/app.log"),
		rotatelogs.WithMaxAge(24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize rotatelogs: %v\n", err)
	}
}

func (l *logger) log(level slog.Level, msg string, data map[string]any) {
	if level < l.level {
		return
	}

	logMessage := logMessage{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level.String(),
		Message:   msg,
		Data:      data,
	}

	logData, err := json.Marshal(logMessage)
	if err != nil {
		fmt.Println("Error marshaling log message:", err)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return
	}

	_, err = l.file.Write(logData)
	if err != nil {
		return
	}

	_, err = l.file.Write([]byte("\n"))
	if err != nil {
		return
	}

	l.file.Sync()
}

func (l *logger) SetFile(filename string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		err := l.file.Close()
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	l.file = file
	l.output = filename
	return nil
}

func (l *logger) SetLevel(level slog.Level) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.level = level
}

func SetLevel(level slog.Level) {
	logInstance.SetLevel(level)
}

func SetFile(filename string) error {
	return logInstance.SetFile(filename)
}

func ReinitializeForTesting(projectRoot string) error {
	logsDir := projectRoot + "/storage/logs"

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	logFilename := fmt.Sprintf("%s/app.%s.log", logsDir, time.Now().Format("2006-01-02"))

	if err := logInstance.SetFile(logFilename); err != nil {
		return fmt.Errorf("failed to set log file: %w", err)
	}

	return nil
}

func Debug(msg string, data ...map[string]any) {
	var logData map[string]any
	if len(data) > 0 {
		logData = data[0]
	}
	logInstance.log(slog.LevelDebug, msg, logData)
}

func Info(msg string, data ...map[string]any) {
	var logData map[string]any
	if len(data) > 0 {
		logData = data[0]
	}
	logInstance.log(slog.LevelInfo, msg, logData)
}

func Warn(msg string, data ...map[string]any) {
	var logData map[string]any
	if len(data) > 0 {
		logData = data[0]
	}
	logInstance.log(slog.LevelWarn, msg, logData)
}

func Error(msg string, data ...map[string]any) {
	var logData map[string]any
	if len(data) > 0 {
		logData = data[0]
	}
	logInstance.log(slog.LevelError, msg, logData)
}

func Fatal(msg string, data ...map[string]any) {
	var logData map[string]any
	if len(data) > 0 {
		logData = data[0]
	}
	logInstance.log(slog.LevelError, msg, logData)

	fmt.Fprintf(os.Stderr, "FATAL ERROR: %s\n", msg)
	if logData != nil && len(logData) > 0 {
		fmt.Fprintf(os.Stderr, "📋 Details:\n")
		for key, value := range logData {
			fmt.Fprintf(os.Stderr, "   %s: %v\n", key, value)
		}
	}

	os.Exit(1)
}

func ParseLogLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level %q", s)
	}
}
