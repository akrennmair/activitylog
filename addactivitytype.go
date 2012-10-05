package main

import (
	"net/http"
	"log"
	"encoding/json"
	"code.google.com/p/gorilla/sessions"
)

type AddActivityTypeHandler struct {
	Store sessions.Store
}

func (h *AddActivityTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	user_id := session.Values["UserId"].(int64)

	typename := r.FormValue("typename")

	result, err := db.Exec("INSERT INTO activity_types (name, user_id, active) VALUES (?, ?, 1)", typename, user_id)
	if err != nil {
		log.Printf("db.Exec failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	activity_type := ActivityType{Name: typename}
	activity_type.Id, _ = result.LastInsertId()

	if json_data, err := json.Marshal(activity_type); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
