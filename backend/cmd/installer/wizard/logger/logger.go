package logger

import (
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()

	logFile := "log.json"
	if envLogFile, ok := os.LookupEnv("INSTALLER_LOG_FILE"); ok {
		logFile = envLogFile
	}

	out, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}

	log.Out = out
	log.Formatter = &logrus.TextFormatter{
		ForceColors:     true,
		DisableQuote:    true,
		TimestampFormat: time.TimeOnly,
	}
}

func Log(message string, args ...any) {
	if log == nil {
		logrus.Infof(message, args...)
		return
	}
	if len(args) > 0 {
		log.Infof(message, args...)
	} else {
		log.Info(message)
	}
}

func Errorf(message string, args ...any) {
	if log == nil {
		logrus.Errorf(message, args...)
		return
	}
	if len(args) > 0 {
		log.Errorf(message, args...)
	} else {
		log.Error(message)
	}
}

func Debugf(message string, args ...any) {
	if log == nil {
		logrus.Debugf(message, args...)
		return
	}
	if len(args) > 0 {
		log.Debugf(message, args...)
	} else {
		log.Debug(message)
	}
}

func Warnf(message string, args ...any) {
	if log == nil {
		logrus.Warnf(message, args...)
		return
	}
	if len(args) > 0 {
		log.Warnf(message, args...)
	} else {
		log.Warn(message)
	}
}

func Fatalf(message string, args ...any) {
	if log == nil {
		logrus.Fatalf(message, args...)
		return
	}
	if len(args) > 0 {
		log.Fatalf(message, args...)
	} else {
		log.Fatal(message)
	}
}

func Panicf(message string, args ...any) {
	if log == nil {
		logrus.Panicf(message, args...)
		return
	}
	if len(args) > 0 {
		log.Panicf(message, args...)
	} else {
		log.Panic(message)
	}
}

func GetLevel() logrus.Level {
	if log == nil {
		return logrus.GetLevel()
	}
	return log.GetLevel()
}

func SetLevel(level logrus.Level) {
	if log == nil {
		logrus.SetLevel(level)
		return
	}
	log.SetLevel(level)
}

func SetOutput(output io.Writer) {
	if log == nil {
		logrus.SetOutput(output)
		return
	}
	log.SetOutput(output)
}
