package logger

import "github.com/sirupsen/logrus"

func Infof(format string, args ...any) {
	logrus.SetFormatter(FormatConfig())
	logrus.Infof(format, args...)
}

func Errorf(format string, args ...any) {
	logrus.SetFormatter(FormatConfig())
	logrus.Errorf(format, args...)
}

func FormatConfig() logrus.Formatter {
	return &logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
}
