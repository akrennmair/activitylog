package main

import (
	"code.google.com/p/gorilla/sessions"
	"encoding/json"
	"log"
	"net/http"
)

type TryAuthenticateHandler struct {
	Db    ActivityTypesGetter
	Store sessions.Store
}

func (h *TryAuthenticateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var result AuthResult
	session, _ := h.Store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		result.Authenticated = false
		result.ErrorMsg = "Authentication failed."
	} else {
		if user_id, ok := session.Values["UserId"].(int64); ok {
			result.Authenticated = true
			result.Activities = h.Db.GetActivityTypesForUser(user_id)
		} else {
			result.Authenticated = false
			result.ErrorMsg = "Authentication failed."
		}
	}

	log.Printf("TryAuthenticate: %v", result)

	if json_data, err := json.Marshal(result); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
