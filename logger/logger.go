package logger

import (
	"io"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	file, err := rotatelogs.New(
		"./storage/logs/app.%Y-%m-%d.log",
		rotatelogs.WithLinkName("./storage/logs/app.log"),
		rotatelogs.WithMaxAge(24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err == nil {
		Logger.SetOutput(file)
	}
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: false,
	})

}

func SetOutput(output io.Writer) {
	Logger.SetOutput(output)
}

func SetLevel(level logrus.Level) {
	Logger.SetLevel(level)
}

func SetFormatter(formatter logrus.Formatter) {
	Logger.SetFormatter(formatter)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return Logger.WithField(key, value)
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

func Warning(args ...interface{}) {
	Logger.Warning(args...)
}

func Warningf(format string, args ...interface{}) {
	Logger.Warningf(format, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	Logger.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	Logger.Panicf(format, args...)
}

type LogRequest struct {
	Logger *logrus.Entry
}

func NewLogRequestWithUUID(c *logrus.Entry, requestUUID string) *LogRequest {
	return &LogRequest{
		Logger: c.WithField("request_uuid", requestUUID),
	}
}
