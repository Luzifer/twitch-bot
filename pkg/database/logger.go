package database

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

type (
	logWriter struct{ io.Writer }
)

func newLogrusLogWriterWithLevel(level logrus.Level, dbDriver string) logWriter {
	writer := logrus.WithField("database", dbDriver).WriterLevel(level)
	return logWriter{writer}
}

func (l logWriter) Print(a ...any) {
	fmt.Fprint(l.Writer, a...)
}

func (l logWriter) Printf(format string, a ...any) {
	fmt.Fprintf(l.Writer, format, a...)
}
