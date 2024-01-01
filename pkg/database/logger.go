package database

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

type (
	// LogWriter implements a logger for the gorm logging
	LogWriter struct{ io.Writer }
)

// NewLogrusLogWriterWithLevel creates a new LogWriter with the given
// logrus.Logger and the specified logrus.Level
func NewLogrusLogWriterWithLevel(logger *logrus.Logger, level logrus.Level, dbDriver string) LogWriter {
	writer := logger.WithField("database", dbDriver).WriterLevel(level)
	return LogWriter{writer}
}

// Print implements the gorm.Logger interface
func (l LogWriter) Print(a ...any) {
	fmt.Fprint(l.Writer, a...)
}

// Printf implements the gorm.Logger interface
func (l LogWriter) Printf(format string, a ...any) {
	fmt.Fprintf(l.Writer, format, a...)
}
