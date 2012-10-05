package main

import (
	"encoding/json"
	"net/http"
)

type ListActivityTypesHandler struct {
	Db ActivityTypesGetter
}

func (h *ListActivityTypesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION_NAME)
	activity_types := []ActivityType{}
	if user_id, ok := session.Values["UserId"].(int64); ok {
		activity_types = h.Db.GetActivityTypesForUser(user_id)
	}

	if json_data, err := json.Marshal(activity_types); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
