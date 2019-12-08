package session

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Tracef(format string, args ...interface{})
}

type Fields map[string]interface{}

type StructuredLogger interface {
	Logger
	WithFields(fields map[string]interface{}) Logger
}

type LogrusLogger struct {
	logger *logrus.Logger
}

func NewLogrusLogger(logger *logrus.Logger) StructuredLogger {
	return &LogrusLogger{logger}
}

func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *LogrusLogger) Tracef(format string, args ...interface{}) {
	l.logger.Tracef(format, args...)
}

func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	return l.logger.WithFields(fields)
}
