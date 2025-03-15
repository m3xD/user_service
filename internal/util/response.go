package util

import (
	"encoding/json"
	"net/http"
)

type ErrReason struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ResponseError struct {
	Status    string      `json:"status"`
	TimeStamp string      `json:"timestamp"`
	Message   string      `json:"message"`
	Errors    []ErrReason `json:"errors,omitempty"`
	Path      string      `json:"path,omitempty"`
}

type ResponseSuccess struct {
	Message string `json:"message"`
}

func ResponseOK(w http.ResponseWriter, res interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), statusCode)
	}
}

func ResponseErr(w http.ResponseWriter, res ResponseError, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), statusCode)
	}
}
