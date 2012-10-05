package main

import (
	"code.google.com/p/gorilla/sessions"
	"net/http"
	"strconv"
	"log"
	"fmt"
)

type DeleteActivityTypeHandler struct {
	Store sessions.Store
	Db ActivityTypeDeleter
}

func (h *DeleteActivityTypeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	user_id := session.Values["UserId"].(int64)

	activity_type_id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)

	log.Printf("delete activity type id %d", activity_type_id);

	if err := h.Db.DeleteActivityType(user_id, activity_type_id); err != nil {
		log.Printf("deactivating activity type %d failed: %v", activity_type_id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: return information about deleted element
	fmt.Fprintf(w, "OK")
}
