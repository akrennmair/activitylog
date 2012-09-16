package main

import (
	"code.google.com/p/gorilla/sessions"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Activity struct {
	Id          string `json:"id"`
	Timestamp   string `json:"ts"`
	Description string `json:"desc"`
}

var (
	activities  map[string][]Activity
	store       sessions.Store
	session_key string = "foobar-supersecret"
)

const (
	ActivityLimit = 10
	SESSION_NAME  = "activities-session"
)

func AddActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "can't anything other than POST", http.StatusMethodNotAllowed)
		return
	}

	session, _ := store.Get(r, SESSION_NAME)
	if !(session.Values["Authenticated"].(bool)) {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	username := session.Values["User"].(string)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "couldn't parse form", http.StatusInternalServerError)
		return
	}

	id := r.FormValue("id")
	description := r.FormValue("desc")
	ts := time.Now().Format(time.RFC3339)

	log.Printf("added activity %s (id = %s) for user %s", description, id, username)
	activities[username] = append([]Activity{{Id: id, Description: description, Timestamp: ts}}, activities[username]...)

	fmt.Fprintf(w, "OK")
}

func LatestActivities(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION_NAME)
	if !(session.Values["Authenticated"].(bool)) {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	username := session.Values["User"].(string)

	max_idx := ActivityLimit
	if len(activities[username]) < ActivityLimit {
		max_idx = len(activities[username])
	}

	if json_data, err := json.Marshal(activities[username][0:max_idx]); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type AuthResult struct {
	Authenticated bool     `json:"authenticated"`
	ErrorMsg      string   `json:"errormsg,omitempty"`
	Activities    []string `json:"activities,omitempty"`
}

func VerifyCredentials(username, password string) bool {
	users := map[string]string{
		"ak":  "foobar",
		"foo": "quux",
	}

	for u, p := range users {
		if u == username && p == password {
			return true
		}
	}

	return false
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "can't anything other than POST", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "couldn't parse form", http.StatusInternalServerError)
	}

	username := r.FormValue("username")
	password := r.FormValue("password")


	var result AuthResult
	if VerifyCredentials(username, password) {
		result.Authenticated = true
		result.Activities = []string{"Eat", "Sleep", "Drink", "Shopping"}

		// create new session and store that authentication was successful
		session, _ := store.Get(r, SESSION_NAME)
		session.Values["Authenticated"] = true
		session.Values["User"] = username
		session.Save(r, w)
	} else {
		result.Authenticated = false
		result.ErrorMsg = "Authentication failed."
	}


	if json_data, err := json.Marshal(result); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	store = sessions.NewCookieStore([]byte(session_key))

	activities = make(map[string][]Activity)

	servemux := http.NewServeMux()

	servemux.Handle("/", http.FileServer(http.Dir("htdocs")))
	servemux.Handle("/auth", http.HandlerFunc(Authenticate))
	servemux.Handle("/activity/add", http.HandlerFunc(AddActivity))
	servemux.Handle("/activity/latest", http.HandlerFunc(LatestActivities))

	httpsrv := &http.Server{Handler: servemux, Addr: ":8000"}
	if err := httpsrv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
