package database

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

type (
	logWriter struct{ io.Writer }
)

func newLogrusLogWriterWithLevel(level logrus.Level) logWriter {
	writer := logrus.StandardLogger().WriterLevel(level)
	return logWriter{writer}
}

func (l logWriter) Printf(format string, a ...any) {
	fmt.Fprintf(l.Writer, format, a...)
}
