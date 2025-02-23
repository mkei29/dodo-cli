package main

import (
	"fmt"

	"github.com/caarlos0/log"
	"github.com/goccy/go-yaml/ast"
)

type AppError struct {
	message string
}

func NewAppError(message string) *AppError {
	return &AppError{
		message: message,
	}
}

func (e *AppError) Error() string {
	return e.message
}

type MultiError struct {
	errors []error
}

func NewMultiError() MultiError {
	return MultiError{
		errors: []error{},
	}
}

func (e *MultiError) Error() string {
	message := fmt.Sprintf("%d errors: ", len(e.errors))
	for _, err := range e.errors {
		message += err.Error() + ", "
	}
	return message
}

func (e *MultiError) Errors() []error {
	return e.errors
}

func (e *MultiError) Add(err error) {
	e.errors = append(e.errors, err)
}

func (e *MultiError) Merge(errs MultiError) {
	e.errors = append(e.errors, errs.errors...)
}

func (e *MultiError) HasError() bool {
	return len(e.errors) > 0
}

func (e *MultiError) Length() int {
	return len(e.errors)
}

func (e *MultiError) Summary() {
	for _, err := range e.errors {
		log.Error(err.Error())
	}
}

type ParseError struct {
	filepath string
	message  string
	line     string
	node     ast.Node
}

func (e *ParseError) Error() string {
	line := e.node.GetToken().Position.Line
	return fmt.Sprintf("%s:%d: %s", e.filepath, line, e.message)
}

func (e *ParseError) Is(target error) bool {
	_, ok := target.(*ParseError)
	return ok
}
