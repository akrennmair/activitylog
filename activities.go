package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Activity struct {
	Id string `json:"id"`
	Timestamp string `json:"ts"`
	Description string `json:"desc"`
}

var (
	activities []Activity
)

const (
	ActivityLimit = 10
)

func AddActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "can't anything other than POST", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "couldn't parse form", http.StatusInternalServerError)
	}

	id := r.FormValue("id")
	description := r.FormValue("desc")
	ts := time.Now().Format(time.RFC3339)

	activities = append([]Activity{{Id: id, Description: description, Timestamp: ts}}, activities...)

	fmt.Fprintf(w, "OK")
}

func LatestActivities(w http.ResponseWriter, r *http.Request) {
	max_idx := ActivityLimit
	if len(activities) < ActivityLimit {
		max_idx = len(activities)
	}

	if json_data, err := json.Marshal(activities[0:max_idx]); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	activities = []Activity{}

	servemux := http.NewServeMux()

	servemux.Handle("/", http.FileServer(http.Dir("htdocs")))
	servemux.Handle("/activity/add", http.HandlerFunc(AddActivity))
	servemux.Handle("/activity/latest", http.HandlerFunc(LatestActivities))

	httpsrv := &http.Server{Handler: servemux, Addr: ":8000"}
	if err := httpsrv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
