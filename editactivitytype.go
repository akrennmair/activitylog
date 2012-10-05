package main

import (
	"code.google.com/p/gorilla/sessions"
	"net/http"
	"log"
	"strconv"
	"fmt"
)

type EditActivityTypeHandler struct {
	Store sessions.Store
}

func (h *EditActivityTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	//user_id := session.Values["UserId"].(int64)

	activity_type_id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)

	// TODO: implement

	new_name := r.FormValue("newname")

	log.Printf("edit activity type id %d new_name = %s", activity_type_id, new_name)

	// TODO: return information about updated activity type

	fmt.Fprintf(w, "OK")
}
