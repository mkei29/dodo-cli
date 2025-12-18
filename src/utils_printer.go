package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	appErrors "github.com/toritoritori29/dodo-cli/src/errors"
)

const (
	NoColor = iota
	ErrorLevel
)

var Styles = [...]StyleFormat{ //nolint: gochecknoglobals
	NoColor: {
		Primary:   lipgloss.NewStyle(),
		Secondary: lipgloss.NewStyle(),
	},
	ErrorLevel: {
		Primary:   lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
		Secondary: lipgloss.NewStyle().Foreground(lipgloss.Color("11")),
	},
}

type PrinterConfig interface {
	EnableColor() bool
	EnablePrinter() bool
}

type StyleFormat struct {
	Primary   lipgloss.Style
	Secondary lipgloss.Style
}

type ErrorPrinter struct {
	stderr  io.Writer
	padding int
	style   StyleFormat
}

func NewErrorPrinter(styleIdx int) *ErrorPrinter {
	return &ErrorPrinter{
		stderr:  os.Stderr,
		padding: 3,
		style:   Styles[styleIdx],
	}
}

func NewPrinterFromArgs(args PrinterConfig) *ErrorPrinter {
	printer := NewErrorPrinter(ErrorLevel)
	if !args.EnableColor() {
		printer.SetStyle(NoColor)
	}
	if !args.EnablePrinter() {
		printer.Disable()
	}
	return printer
}

func (p *ErrorPrinter) Disable() {
	p.stderr = io.Discard
}

func (p *ErrorPrinter) SetStyle(styleIdx int) {
	p.style = Styles[styleIdx]
}

// PrettyPrint prints the error in a human-readable format.
func (p *ErrorPrinter) HandleError(err error) error {
	// if the error is a MultiError, call PrettyPrint on each error.
	// MultiError doesn't implement Unwrap, so we can't use errors.Is.

	if errors.Is(err, appErrors.ErrAlreadyHandled) {
		return err
	}

	var merr *appErrors.MultiError
	if errors.As(err, &merr) {
		for _, e := range merr.Errors() {
			p.HandleError(e) //nolint: errcheck
		}
		return appErrors.ErrAlreadyHandled
	}

	var perr *appErrors.ParseError
	if errors.As(err, &perr) {
		p.printParseError(perr)
		return appErrors.ErrAlreadyHandled
	}
	p.printError(err)
	return appErrors.ErrAlreadyHandled
}

func (p *ErrorPrinter) printParseError(err *appErrors.ParseError) {
	// Print a parse error in a human-readable format.
	// This function respects the golangci-lint style error format.
	//
	// e.g.
	// error.go:123:1: failed to parse time: failed to parse time: parsing time
	//     time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")

	listIcon := p.style.Primary.Render(fmt.Sprintf("%*s", p.padding, "⨯"))
	line := err.Node.GetToken().Position.Line
	pos := err.Node.GetToken().Position.Column
	message := p.style.Primary.Render(err.Message)
	fmt.Fprintf(p.stderr, "%s %s:%d:%d %s\n", listIcon, err.Filepath, line, pos, message)

	arrow := p.style.Secondary.Render(fmt.Sprintf("%*s", p.padding+2, ">"))
	fmt.Fprintf(p.stderr, "%s %s\n", arrow, err.Line)
}

func (p *ErrorPrinter) printError(err error) {
	// Print a general error in a human-readable format.
	// This function respects the golangci-lint style error format.
	listIcon := p.style.Primary.Render(fmt.Sprintf("%*s", p.padding, "⨯"))
	message := p.style.Primary.Render(err.Error())
	fmt.Fprintf(p.stderr, "%s %s\n", listIcon, message)
}
