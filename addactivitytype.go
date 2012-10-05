package main

import (
	"net/http"
	"encoding/json"
	"code.google.com/p/gorilla/sessions"
)

type AddActivityTypeHandler struct {
	Store sessions.Store
	Db ActivityTypeAdder
}

func (h *AddActivityTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	user_id := session.Values["UserId"].(int64)
	typename := r.FormValue("typename")

	activity_type, err := h.Db.AddActivityType(typename, user_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if json_data, err := json.Marshal(activity_type); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
