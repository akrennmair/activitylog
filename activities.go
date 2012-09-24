package main

import (
	"bytes"
	"code.google.com/p/go.crypto/pbkdf2"
	"code.google.com/p/gorilla/sessions"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"time"
)

type Activity struct {
	TypeId      string `json:"type_id" bson:"type_id"`
	Timestamp   string `json:"ts"`
	Description string `json:"desc"`
	User        string `json:"-"`
}

var (
	store       sessions.Store
	session_key string = "foobar-supersecret"
	db          *mgo.Database
)

const (
	ActivityLimit      = 10
	SESSION_NAME       = "activities-session"
	COLL_ACTIVITY      = "activity"
	COLL_USERS         = "users"
	COLL_ACTIVITY_TYPE = "activity_types"
	DB_NAME            = "activitylog"
	PBKDF2_ROUNDS      = 10000
	PBKDF2_SIZE        = 32
)

type ActivityType struct {
	_Id  bson.ObjectId `bson:"_id" json:"-"`
	Id   string        `json:"type_id" bson:"-"`
	Name string        `json:"name"`
	User string        `json:"user,omitempty"`
}

func AddActivityType(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "can't anything other than POST", http.StatusMethodNotAllowed)
		return
	}

	session, _ := store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	username := session.Values["User"].(string)

	if err := r.ParseForm(); err != nil {
		log.Printf("r.ParseForm failed: %v", err)
		http.Error(w, "couldn't parse form", http.StatusInternalServerError)
		return
	}

	typename := r.FormValue("typename")


	object_id := bson.NewObjectId()
	activity_type := bson.M{"_id": object_id, "name": typename, "user": username}
	if err := db.C(COLL_ACTIVITY_TYPE).Insert(&activity_type); err != nil {
		log.Printf("c.Insert failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	activity_type_json := ActivityType{Id: object_id.Hex(), Name: typename, User: username}

	log.Printf("activity_type after insert: %#v", activity_type_json)

	if json_data, err := json.Marshal(activity_type_json); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func AddActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "can't anything other than POST", http.StatusMethodNotAllowed)
		return
	}

	session, _ := store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	username := session.Values["User"].(string)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "couldn't parse form", http.StatusInternalServerError)
		return
	}

	type_id := r.FormValue("type_id")
	description := r.FormValue("desc")
	ts := time.Now().Format(time.RFC3339)

	log.Printf("added activity %s (type_id = %s) for user %s", description, type_id, username)
	activity := Activity{TypeId: type_id, Description: description, Timestamp: ts, User: username}
	if err := db.C(COLL_ACTIVITY).Insert(&activity); err != nil {
		log.Printf("c.Insert failed: %v", err)
	}

	fmt.Fprintf(w, "OK")
}

func LatestActivities(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	username := session.Values["User"].(string)

	var activities []Activity
	if err := db.C(COLL_ACTIVITY).Find(bson.M{"user": username}).Sort("-timestamp", "-_id").Limit(ActivityLimit).All(&activities); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if json_data, err := json.Marshal(activities); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type AuthResult struct {
	Authenticated bool           `json:"authenticated"`
	ErrorMsg      string         `json:"errormsg,omitempty"`
	Activities    []ActivityType `json:"activities,omitempty"`
}

func VerifyCredentials(username, password string) bool {
	var user User
	if err := db.C(COLL_USERS).Find(bson.M{"_id": username}).One(&user); err != nil {
		log.Printf("finding user %s failed: %v", username, err)
		return false
	}

	password_hash := pbkdf2.Key([]byte(password), user.Salt, PBKDF2_ROUNDS, PBKDF2_SIZE, sha256.New)

	return bytes.Equal(password_hash, user.Password)
}

func GenerateSalt() (data []byte, err error) {
	data = make([]byte, 8)
	_, err = rand.Read(data)
	return
}

type User struct {
	Id       string `_id`
	Password []byte
	Salt     []byte
}

func RegisterUser(username, password string) error {
	salt, err := GenerateSalt()
	if err != nil {
		return err
	}

	password_hash := pbkdf2.Key([]byte(password), salt, PBKDF2_ROUNDS, PBKDF2_SIZE, sha256.New)

	new_user := &User{Id: username, Password: password_hash, Salt: salt}
	if err = db.C(COLL_USERS).Insert(&new_user); err != nil {
		return err
	}

	return nil
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "can't anything other than POST", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "couldn't parse form", http.StatusInternalServerError)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var result AuthResult

	if err := RegisterUser(username, password); err != nil {
		result.Authenticated = false
		result.ErrorMsg = err.Error()
	} else {
		result.Authenticated = true
	}

	if json_data, err := json.Marshal(result); err == nil {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json_data)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "can't anything other than POST", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "couldn't parse form", http.StatusInternalServerError)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var result AuthResult
	if VerifyCredentials(username, password) {
		result.Authenticated = true

		if err := db.C(COLL_ACTIVITY_TYPE).Find(bson.M{"user": username}).All(&result.Activities); err != nil {
			log.Printf("Find failed: %v", err)
		}

		for i, _ := range result.Activities {
			result.Activities[i].Id = result.Activities[i]._Id.Hex()
		}

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
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatalf("mgo.Dial: %v", err)
	}
	defer session.Close()

	db = session.DB(DB_NAME)

	store = sessions.NewCookieStore([]byte(session_key))

	servemux := http.NewServeMux()

	servemux.Handle("/", http.FileServer(http.Dir("htdocs")))
	servemux.Handle("/auth", http.HandlerFunc(Authenticate))
	servemux.Handle("/auth/signup", http.HandlerFunc(Signup))
	servemux.Handle("/activity/add", http.HandlerFunc(AddActivity))
	servemux.Handle("/activity/latest", http.HandlerFunc(LatestActivities))
	servemux.Handle("/activity/type/add", http.HandlerFunc(AddActivityType))

	httpsrv := &http.Server{Handler: servemux, Addr: ":8000"}
	if err := httpsrv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
