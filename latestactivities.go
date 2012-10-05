package main

import (
	"net/http"
	"encoding/json"
)

type LatestActivitiesHandler struct {
}

func (h *LatestActivitiesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION_NAME)
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
