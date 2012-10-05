package main

import (
	"net/http"
	"fmt"
)

type LogoutHandler struct {
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body.Close()

	session, _ := store.Get(r, SESSION_NAME)
	delete(session.Values, "Authenticated")
	delete(session.Values, "UserName")
	delete(session.Values, "UserId")
	session.Save(r, w)

	// TODO: maybe print JSON?
	fmt.Fprintf(w, "OK")
}
