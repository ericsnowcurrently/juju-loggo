// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package loggo

// Formatter defines the single method Format, which takes the logging
// record and converts it to a string.
type Formatter interface {
	Format(Record) string
}

// MinimalFormatter is a formatter that produces only the message.
type MinimalFormatter struct{}

func (*MinimalFormatter) Format(rec Record) string {
	return rec.Message
}

// BasicFormatter is a simple formatter that produces something like:
//   WARNING The message...
type BasicFormatter struct{}

func (*BasicFormatter) Format(rec Record) string {
	return fmt.Sprintf("%s %s", rec.Level, rec.Message)
}
