package main

type AppError struct {
	message string
}

func (e *AppError) Error() string {
	return e.message
}
