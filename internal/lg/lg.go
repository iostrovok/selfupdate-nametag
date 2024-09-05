package lg

import (
	"os"

	"github.com/sirupsen/logrus"
)

/*
	lg is a simple wrapper around logrus
	todo: rotation, log levels, etc
*/

type Logger struct {
	lg      *logrus.Logger
	version string
}

func New(logFile, version string) (*Logger, error) {
	var log = logrus.New()
	log.Formatter = new(logrus.TextFormatter) //default
	log.Level = logrus.TraceLevel

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Info("Failed to log to file, using default stderr")
		return nil, err
	}

	log.Out = file
	return &Logger{lg: log}, nil
}

func (l *Logger) L() *logrus.Logger {
	return l.lg
}

// Close just helps the garbage collector
func (l *Logger) Close() {
	l.lg.Formatter = nil
	l.lg.Hooks = nil
	l.lg.ExitFunc = nil
	l.lg.Out = nil
	l.lg = nil
}

func (l *Logger) Errorf(format string, args ...any) {
	l.lg.WithField("version", l.version).Errorf(format, args...)
}

func (l *Logger) Fatal(args ...any) {
	l.lg.WithField("version", l.version).Fatal(args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.lg.WithField("version", l.version).Fatalf(format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.lg.WithField("version", l.version).Infof(format, args...)
}

func (l *Logger) Printf(format string, args ...any) {
	l.lg.WithField("version", l.version).Printf(format, args...)
}
