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

// ModuleWriter only writes log messages for the named module.
type ModuleWriter struct {
	RecordWriter

	// Name is the module name to look for.
	Name string
}

// Write writes the record to the wrapped writer, but only if the module
// name matches.
func (w *ModuleWriter) WriteRecord(rec Record) {
	if rec.Module == w.Name {
		w.RecordWriter.Write(rec)
	}
}

// StdioWriter writes log messages at WARNING, ERROR, and CRITICAL to
// a stderr writer. The rest go to a stdout writer.
// LogWriter filters the log messages for name.
type StdioWriter struct {
	// Out is the writer to use for stdout.
	Out RecordWriter

	// Err is the writer to use for stderr.
	Err RecordWriter
}

// NewStdioWriter returns a new StdioWriter with formatting writers
// wrapping the provided io.Writers.
func NewStdioWriter(out, err io.Writer, formatter Formatter) *StdioWriter {
	if formatter == nil {
		formatter = &MinimalFormatter{}
	}
	return &StdioWriter{
		Out: NewFormattingWriter(out, formatter),
		Err: NewFormattingWriter(err, formatter),
	}
}

// WriteRecord writes to the stdout writer if the log level is INFO
// or below. The stderr writer is used otherwise.
func (w *StdioWriter) WriteRecord(rec Record) {
	if rec.Level <= INFO {
		w.Out.WriteRecord(rec)
	} else {
		w.Err.WriteRecord(rec)
	}
}

// DiscardWriter throws away records.
type DiscardWriter struct{}

// WriteRecord is a no-op.
func (DiscardWriter) WriteRecord(Record) {
}

type ChannelWriter struct {
	Records <-chan Record
}

func (w ChannelWriter) WriteRecord(rec Record) {
	select {
	case w.Records = <-rec:
	default:
	}
}

type MemoryWriter struct {
	Records interface {
		Add(Record)
	}
}

// WriteRecord stores the record in memory.
func (w *MemoryWriter) WriteRecord(rec Record) {
	w.Records.Add(rec)
}
