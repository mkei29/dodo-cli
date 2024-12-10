package main

import (
	"github.com/caarlos0/log"
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

type ErrorSet struct {
	errors []error
}

func NewErrorSet() ErrorSet {
	return ErrorSet{
		errors: []error{},
	}
}

func (e *ErrorSet) Errors() []error {
	return e.errors
}

func (e *ErrorSet) Add(err error) {
	e.errors = append(e.errors, err)
}

func (e *ErrorSet) Merge(errs ErrorSet) {
	e.errors = append(e.errors, errs.errors...)
}

func (e *ErrorSet) HasError() bool {
	return len(e.errors) > 0
}

func (e *ErrorSet) Length() int {
	return len(e.errors)
}

func (e *ErrorSet) Log() {
	for _, err := range e.errors {
		log.Error(err.Error())
	}
}

func (e *ErrorSet) Summary() {
	for _, err := range e.errors {
		log.Debug(err.Error())
	}
}
