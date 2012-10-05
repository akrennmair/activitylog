package main

import (
	"code.google.com/p/gorilla/sessions"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type AddActivityHandler struct {
	Db    ActivityAdder
	Store sessions.Store
}

func (h *AddActivityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	username := session.Values["UserName"].(string)
	user_id := session.Values["UserId"].(int64)

	type_id, _ := strconv.ParseInt(r.FormValue("type_id"), 10, 64)
	description := r.FormValue("desc")
	is_public := false
	if r.FormValue("public") == "on" {
		is_public = true
	}

	if err := h.Db.AddActivity(type_id, description, user_id, is_public, r.FormValue("lat"), r.FormValue("long")); err != nil {
		log.Printf("AddActivity failed: %v", err)
	} else {
		log.Printf("added activity %s (type_id = %s) for user %s", description, type_id, username)
	}

	// TODO: return inserted data as JSON including insert ID
	fmt.Fprintf(w, "OK")
}
