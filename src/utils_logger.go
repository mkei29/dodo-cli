package main

import (
	"errors"
	"io"

	"github.com/caarlos0/log"
	"github.com/charmbracelet/lipgloss"
)

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

var NoColorStyles = [...]lipgloss.Style{ //nolint: gochecknoglobals
	LogLevelDebug: lipgloss.NewStyle(),
	LogLevelInfo:  lipgloss.NewStyle(),
	LogLevelWarn:  lipgloss.NewStyle(),
	LogLevelError: lipgloss.NewStyle(),
	LogLevelFatal: lipgloss.NewStyle(),
}

type LoggingConfig interface {
	DisableLogging() bool
	EnableDebugMode() bool
	EnableColor() bool
}

func InitLogger(config LoggingConfig) error {
	if config.DisableLogging() {
		logger, ok := log.Log.(*log.Logger)
		if !ok {
			return errors.New("failed to cast logger to *log.Logger")
		}
		logger.Writer = io.Discard
		return nil
	}

	log.SetLevel(log.InfoLevel)
	if config.EnableDebugMode() {
		log.SetLevel(log.DebugLevel)
		log.Debug("running in debug mode")
	}

	if !config.EnableColor() {
		log.Styles = NoColorStyles
	}
	return nil
}
