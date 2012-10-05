package main

import (
	"net/http"
	"encoding/json"
)

type AuthenticateHandler struct {
	Db ActivityTypesGetter
}

func (h *AuthenticateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	var result AuthResult
	if user_id, ok := VerifyCredentials(username, password); ok {
		result.Authenticated = true
		result.Activities = h.Db.GetActivityTypesForUser(user_id)

		// create new session and store that authentication was successful
		session, _ := store.Get(r, SESSION_NAME)
		session.Values["Authenticated"] = true
		session.Values["UserName"] = username
		session.Values["UserId"] = user_id
		session.Save(r, w)
	} else {
		result.Authenticated = false
		result.ErrorMsg = "Authentication failed."
	}

	if json_data, err := json.Marshal(result); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
