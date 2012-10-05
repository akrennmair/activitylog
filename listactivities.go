package main

import (
	"code.google.com/p/gorilla/sessions"
	"encoding/json"
	"net/http"
	"strconv"
)

type ListActivitiesHandler struct {
	Store sessions.Store
	Db    *Database
}

func (h *ListActivitiesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	user_id := session.Values["UserId"].(int64)

	page, err := strconv.ParseInt(r.URL.Query().Get(":page"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	activities, err := h.Db.GetActivitiesForUser(user_id, ActivityLimit, uint(page-1)*ActivityLimit)
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
