package main

import (
	"net/http"
	"strconv"
	"log"
	"fmt"
)

type DeleteActivityTypeHandler struct {
}

func (h *DeleteActivityTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	user_id := session.Values["UserId"].(int64)

	activity_type_id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)

	log.Printf("delete activity type id %d", activity_type_id);

	_, err := db.Exec("UPDATE activity_types SET active = 0 WHERE user_id = ? AND id = ?", user_id, activity_type_id);
	if err != nil {
		log.Printf("db.Exec failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: return information about deleted element
	fmt.Fprintf(w, "OK")
}
