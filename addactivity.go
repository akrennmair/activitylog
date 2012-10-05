package main

import (
	"net/http"
	"database/sql"
	"log"
	"fmt"
	"strconv"
	"code.google.com/p/gorilla/sessions"
)

type AddActivityHandler struct {
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

	type_id := r.FormValue("type_id")
	description := r.FormValue("desc")
	is_public := 0
	if r.FormValue("public") == "on" {
		is_public = 1
	}

	var err error
	var latitude sql.NullFloat64
	var longitude sql.NullFloat64

	if latitude.Float64, err = strconv.ParseFloat(r.FormValue("lat"), 64); err != nil {
		latitude.Valid = false
	} else {
		latitude.Valid = true
	}

	if longitude.Float64, err = strconv.ParseFloat(r.FormValue("long"), 64); err != nil {
		longitude.Valid = false
	} else {
		longitude.Valid = true
	}

	if _, err := db.Exec("INSERT INTO activities (type_id, timestamp, description, user_id, public, latitude, longitude) VALUES (?, NOW(), ?, ?, ?, ?, ?)", type_id, description, user_id, is_public, latitude, longitude); err != nil {
		log.Printf("AddActivity: db.Exec failed: %v", err)
	} else {
		log.Printf("added activity %s (type_id = %s) for user %s", description, type_id, username)
	}

	// TODO: return inserted data as JSON including insert ID
	fmt.Fprintf(w, "OK")
}
