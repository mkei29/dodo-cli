package main

import (
	"encoding/json"
	"fmt"
)

const (
	JSONLogLevelDisabled = 1
	JSONLogLevelEnabled  = 2
)

type JSONFormat struct {
	Status      string `json:"status"`
	DocumentURL string `json:"document_url"`
	Error       string `json:"error,omitempty"`
}

type JSONWriter struct {
	level int
}

func NewJSONWriter(level int) *JSONWriter {
	return &JSONWriter{level: level}
}

func (j *JSONWriter) Write(p string) {
	if j.level == JSONLogLevelEnabled {
		fmt.Println(p) //nolint: forbidigo
	}
}

func (j *JSONWriter) ShowSucceededJSONText(documentURL string) {
	data := JSONFormat{
		Status:      "success",
		DocumentURL: documentURL,
	}
	// Convert data to JSON format
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	j.Write(string(jsonData))
}

func (j *JSONWriter) ShowFailedJSONText(err error) {
	data := JSONFormat{
		Status: "failed",
		Error:  err.Error(),
	}
	// Convert data to JSON format
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	j.Write(string(jsonData))
}
