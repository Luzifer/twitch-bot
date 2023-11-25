package database

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

type (
	LogWriter struct{ io.Writer }
)

func NewLogrusLogWriterWithLevel(logger *logrus.Logger, level logrus.Level, dbDriver string) LogWriter {
	writer := logger.WithField("database", dbDriver).WriterLevel(level)
	return LogWriter{writer}
}

func (l LogWriter) Print(a ...any) {
	fmt.Fprint(l.Writer, a...)
}

func (l LogWriter) Printf(format string, a ...any) {
	fmt.Fprintf(l.Writer, format, a...)
}
