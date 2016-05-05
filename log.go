// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package loggo

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Log describes a logging configuration.
type Log struct {
	// If DefaultConfig is set, it will be used for the
	// default logging configuration.
	DefaultConfig string
	Path          string
	Debug         bool
	ShowLog       bool
	Config        string

	// NewWriter creates a new logging writer for a specified target.
	NewWriter func(io.Writer) RecordWriter
}

// writer returns a logging writer for the specified target.
func (log Log) writer(target io.Writer) RecordWriter {
	if log.NewWriter == nil {
		return NewFormattingWriter(target, nil)
	}
	return log.NewWriter(target)
}

// Start starts logging using the given Context.
func (log *Log) Start(ctx *Context) error {
	if log.Path != "" {
		path := ctx.AbsPath(log.Path)
		target, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		writer := log.GetLogWriter(target)
		err = loggo.RegisterWriter("logfile", writer, loggo.TRACE)
		if err != nil {
			return err
		}
	}

	level := loggo.WARNING
	if log.ShowLog {
		level = loggo.INFO
	}
	if log.Debug {
		log.ShowLog = true
		level = loggo.DEBUG
	}

	if log.ShowLog {
		// We replace the default writer to use ctx.Stderr rather than os.Stderr.
		writer := log.GetLogWriter(ctx.Stderr)
		_, err := loggo.ReplaceDefaultWriter(writer)
		if err != nil {
			return err
		}
	} else {
		loggo.RemoveWriter("default")
		// Create a simple writer that doesn't show filenames, or timestamps,
		// and only shows warning or above.
		writer := loggo.NewSimpleWriter(ctx.Stderr, &warningFormatter{})
		err := loggo.RegisterWriter("warning", writer, loggo.WARNING)
		if err != nil {
			return err
		}
	}

	// Set the level on the root logger.
	loggo.GetLogger("").SetLogLevel(level)
	// Override the logging config with specified logging config.
	loggo.ConfigureLoggers(log.Config)
	return nil
}
