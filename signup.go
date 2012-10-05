package main

import (
	"net/http"
	"encoding/json"
)

type SignupHandler struct {
	Db UserRegistrar
}

func (h *SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	var result AuthResult

	if err := h.Db.RegisterUser(username, password); err != nil {
		result.Authenticated = false
		result.ErrorMsg = err.Error()
	} else {
		result.Authenticated = true
	}

	if json_data, err := json.Marshal(result); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
