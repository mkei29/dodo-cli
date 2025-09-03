package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
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

type Printer struct {
	writer  io.Writer
	padding int
	style   StyleFormat
}

func NewPrinter(styleIdx int) *Printer {
	return &Printer{
		writer:  os.Stdout,
		padding: 3,
		style:   Styles[styleIdx],
	}
}

func NewPrinterFromArgs(args PrinterConfig) *Printer {
	printer := NewPrinter(ErrorLevel)
	if !args.EnableColor() {
		printer.SetStyle(NoColor)
	}
	if !args.EnablePrinter() {
		printer.Disable()
	}
	return printer
}

func (p *Printer) Disable() {
	p.writer = io.Discard
}

func (p *Printer) SetStyle(styleIdx int) {
	p.style = Styles[styleIdx]
}

// PrettyPrint prints the error in a human-readable format.
func (p *Printer) HandleError(err error) error {
	// if the error is a MultiError, call PrettyPrint on each error.
	// MultiError doesn't implement Unwrap, so we can't use errors.Is.

	if errors.Is(err, ErrAlreadyHandled) {
		return err
	}

	var merr *MultiError
	if errors.As(err, &merr) {
		for _, e := range merr.Errors() {
			p.HandleError(e) //nolint: errcheck
		}
		return ErrAlreadyHandled
	}

	var perr *ParseError
	if errors.As(err, &perr) {
		p.printParseError(perr)
		return ErrAlreadyHandled
	}
	p.printError(err)
	return ErrAlreadyHandled
}

func (p *Printer) printParseError(err *ParseError) {
	// Print a parse error in a human-readable format.
	// This function respects the golangci-lint style error format.
	//
	// e.g.
	// error.go:123:1: failed to parse time: failed to parse time: parsing time
	//     time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")

	listIcon := p.style.Primary.Render(fmt.Sprintf("%*s", p.padding, "⨯"))
	line := err.node.GetToken().Position.Line
	pos := err.node.GetToken().Position.Column
	message := p.style.Primary.Render(err.message)
	fmt.Fprintf(p.writer, "%s %s:%d:%d %s\n", listIcon, err.filepath, line, pos, message)

	arrow := p.style.Secondary.Render(fmt.Sprintf("%*s", p.padding+2, ">"))
	fmt.Fprintf(p.writer, "%s %s\n", arrow, err.line)
}

func (p *Printer) printError(err error) {
	// Print a general error in a human-readable format.
	// This function respects the golangci-lint style error format.
	listIcon := p.style.Primary.Render(fmt.Sprintf("%*s", p.padding, "⨯"))
	message := p.style.Primary.Render(err.Error())
	fmt.Fprintf(p.writer, "%s %s\n", listIcon, message)
}
