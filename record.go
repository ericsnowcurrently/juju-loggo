// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package loggo

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"
)

// Record holds the information for a single log record.
type Record struct {
	// Level is the log level of the log record.
	Level Level

	// Module is the name of the logger for which the record
	// was generated.
	Module string

	// Filename is the path to the code that triggered the record.
	Filename string

	// Line is the line number of the code that triggered the record.
	Line int

	// Timestamp is the time that the record was created.
	Timestamp time.Time

	// Message is the requested log message.
	Message string
}

// NewRecord creates a new log record for the given log level, module,
// and message. It uses the identified entry the call stack to determine
// the filename and line number. The current time is used for the
// timestamp.
func NewRecord(calldepth int, level Level, module, message string) Record {
	// Gather time, and filename, line number.
	now := time.Now() // get this early.
	// Param to Caller is the call depth.  Since this method is called from
	// the Logger methods, we want the place that those were called from.
	_, file, line, ok := runtime.Caller(calldepth + 1)
	if !ok {
		file = "???"
		line = 0
	}
	// Trim newline off format string, following usual
	// Go logging conventions.
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[0 : len(message)-1]
	}
	return Record{
		Level:     level,
		Module:    module,
		Filename:  file,
		Line:      line,
		Timestamp: now,
		Message:   message,
	}
}

// NewRecordf creates a new log record for the given info. The only
// difference from NewRecord() is that the provided args are applied// to the message using fmt.Sprintf().
func NewRecordf(calldepth int, level Level, module, message string, args ...interface{}) Record {
	rec := NewRecord(calldepth+1, level, module, message)
	if len(args) == 0 {
		return rec
	}

	// To avoid having a proliferation of Info/Infof methods,
	// only use Sprintf if there are any args, and rely on the
	// `go vet` tool for the obvious cases where someone has forgotten
	// to provide an arg.
	rec.Message = fmt.Sprintf(rec.Message, args...)

	return rec
}

// String returns the default string representation of the log record.
// The details are separated by spaces except for filename and line
// which are separated by a colon. The timestamp is shown to second
// resolution in UTC.
func (rec Record) String() string {
	ts := rec.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05")
	// Just get the basename from the filename
	filename := filepath.Base(rec.Filename)
	return fmt.Sprintf("%s %s %s %s:%d %s", ts, rec.Level, rec.Module, filename, rec.Line, rec.Message)
}
