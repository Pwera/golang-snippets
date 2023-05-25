package controller

import (
	"encoding/json"
	"net/http"
)

type BaseController struct{}

func (bc BaseController) MarshalAndWriteHeaders(data interface{}, w http.ResponseWriter) {
	usersJson, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(usersJson)
}
