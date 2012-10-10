package main

import (
	"code.google.com/p/gorilla/sessions"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type EditActivityTypeHandler struct {
	Store sessions.Store
	Db    ActivityTypeRenamer
}

func (h *EditActivityTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	user_id := session.Values["UserId"].(int64)

	activity_type_id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)

	new_name := r.FormValue("newname")

	err := h.Db.RenameActivityType(new_name, user_id, activity_type_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		log.Printf("edit activity type id %d new_name = %s", activity_type_id, new_name)
		// TODO: maybe return JSON
		fmt.Fprintf(w, "OK")
	}
}
