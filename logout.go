package main

import (
	"code.google.com/p/gorilla/sessions"
	"fmt"
	"net/http"
)

type LogoutHandler struct {
	Store sessions.Store
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body.Close()

	session, _ := h.Store.Get(r, SESSION_NAME)
	delete(session.Values, "Authenticated")
	delete(session.Values, "UserName")
	delete(session.Values, "UserId")
	session.Save(r, w)

	// TODO: maybe print JSON?
	fmt.Fprintf(w, "OK")
}
