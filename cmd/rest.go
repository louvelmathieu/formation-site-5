package main

import (
	"encoding/json"
	"net/http"
)

func unauthorized(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusUnauthorized)
	d, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	})
	w.Write(d)
}

func badRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	d, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	})
	w.Write(d)
}

func notFound(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusNotFound)
	d, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	})
	w.Write(d)
}
