// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package loggo

import (
	"fmt"
	"io"
)

// defaultWriterName is the name of the writer default writer.
const defaultWriterName = "default"

// RecordWriter is implemented by any recipient of log messages.
type RecordWriter interface {
	// WriteRecord writes a message to the Writer for the given
	// log record.
	WriteRecord(Record)
}

// MinLevelWriter is a writer that exposes its minimum log level.
type MinLevelWriter interface {
	RecordWriter
	HasMinLevel
}

type minLevelWriter struct {
	writer RecordWriter
	level  Level
}

// NewMinLevelWriter returns a MinLevelWriter that wraps the given
// writer with the provided min log level.
func NewMinLevelWriter(writer RecordWriter, minLevel Level) MinLevelWriter {
	return &minLevelWriter{
		writer: writer,
		level:  minLevel,
	}
}

// MinLogLevel returns the writer's log level.
func (w minLevelWriter) MinLogLevel() Level {
	return w.level
}

// Write writes the log record.
func (w minLevelWriter) WriteRecord(rec Record) {
	if !IsLevelEnabled(&w, rec.Level) {
		return
	}
	w.writer.WriteRecord(rec)
}

// formattingWriter is a log writer that writes
// log messages to the given io.Writer, formatting the
// messages with the given formatter.
type formattingWriter struct {
	writer    io.Writer
	formatter Formatter
}

// NewFormattingWriter returns a new writer that writes
// log messages to the given io.Writer, formatting the
// messages with the given formatter.
func NewFormattingWriter(writer io.Writer, formatter Formatter) RecordWriter {
	return &formattingWriter{
		writer:    writer,
		formatter: formatter,
	}
}

// Write formats the record and writes the result to the io.Writer.
func (fw *formattingWriter) WriteRecord(rec Record) {
	var logLine string
	if fw.formatter == nil {
		logLine = rec.String()
	} else {
		logLine = fw.formatter.Format(rec)
	}
	fmt.Fprintln(fw.writer, logLine)
}

// TeeWriter is a MinLevelWriter that writes to a list of writers,
// in order.
type TeeWriter struct {
	combinedMinLevel Level
	writers          []Writer
}

// NewTeeWriter creates a new TeeWriter that will write to the given
// writers, in the order they were provided.
func NewTeeWriter(writers ...Writer) *TeeWriter {
	tw := &TeeWriter{
		combinedMinLevel: UNSPECIFIED,
		writers:          writers,
	}
	if len(writers) > 0 {
		combinedLevel := CRITICAL
		for _, w := range writers {
			mlw, ok := w.(MinLevelWriter)
			if !ok {
				combinedLevel = UNSPECIFIED
				break
			}
			minLevel := mlw.MinLogLevel()
			if minLevel < combinedLevel {
				combinedLevel = minLevel
			}
		}
		tw.combinedMinLevel = combinedLevel
	}
	return tw
}

// MinLogLevel returns the minimum log level at which at least one of
// the registered writers will write.
func (tw *TeeWriter) MinLogLevel() Level {
	return tw.combinedMinLevel
}

// Write implements Writer, sending the message to each registered writer.
func (tw *TeeWriter) Write(rec Record) {
	for _, w := range tw.writers {
		if mlw, ok := w.(MinLevelWriter); !ok || !IsLevelEnabled(mlw, rec.Level) {
			continue
		}
		w.Write(rec)
	}
}
