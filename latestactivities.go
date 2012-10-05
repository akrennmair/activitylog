package main

import (
	"code.google.com/p/gorilla/sessions"
	"net/http"
	"encoding/json"
)

type LatestActivitiesHandler struct {
	Store sessions.Store
}

func (h *LatestActivitiesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	user_id := session.Values["UserId"].(int64)

	activities, err := GetActivitiesForUser(user_id, ActivityLimit, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if json_data, err := json.Marshal(activities); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
