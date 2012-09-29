package main

import (
	"bytes"
	_ "code.google.com/p/go-mysql-driver/mysql"
	"code.google.com/p/go.crypto/pbkdf2"
	"code.google.com/p/gorilla/sessions"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	goconf "github.com/akrennmair/goconf"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Activity struct {
	TypeId      int64    `json:"type_id"`
	Timestamp   string   `json:"ts"`
	Description string   `json:"desc"`
	Latitude    *float64 `json:"lat"`
	Longitude   *float64 `json:"long"`
}

var (
	store sessions.Store
	db    *sql.DB
)

const (
	ActivityLimit = 10
	SESSION_NAME  = "activities-session"
	DB_NAME       = "activitylog"
	PBKDF2_ROUNDS = 10000
	PBKDF2_SIZE   = 32
)

type ActivityType struct {
	Id   int64  `json:"type_id"`
	Name string `json:"name"`
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

	user_id := session.Values["UserId"].(int64)

	if err := r.ParseForm(); err != nil {
		log.Printf("r.ParseForm failed: %v", err)
		http.Error(w, "couldn't parse form", http.StatusInternalServerError)
		return
	}

	typename := r.FormValue("typename")

	result, err := db.Exec("INSERT INTO activity_types (name, user_id, active) VALUES (?, ?, 1)", typename, user_id)
	if err != nil {
		log.Printf("db.Exec failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	activity_type := ActivityType{Name: typename}
	activity_type.Id, _ = result.LastInsertId()

	if json_data, err := json.Marshal(activity_type); err == nil {
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

	username := session.Values["UserName"].(string)
	user_id := session.Values["UserId"].(int64)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "couldn't parse form", http.StatusInternalServerError)
		return
	}

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

	fmt.Fprintf(w, "OK")
}

func LatestActivities(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, SESSION_NAME)
	if authenticated, ok := session.Values["Authenticated"].(bool); !ok || !authenticated {
		http.Error(w, "unauthenticated", http.StatusForbidden)
		return
	}

	user_id := session.Values["UserId"].(int64)

	rows, err := db.Query("SELECT type_id, timestamp, description, latitude, longitude FROM activities WHERE user_id = ? ORDER BY timestamp DESC LIMIT ?", user_id, ActivityLimit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	activities := []Activity{}

	for rows.Next() {
		var type_id int64
		var timestamp string
		var description string
		var longitude, latitude string
		if err = rows.Scan(&type_id, &timestamp, &description, &latitude, &longitude); err == nil {
			activity := Activity{TypeId: type_id, Timestamp: timestamp, Description: description}
			if latitude != "" {
				activity.Latitude = new(float64)
				*activity.Latitude, _ = strconv.ParseFloat(latitude, 64)
			}
			if longitude != "" {
				activity.Longitude = new(float64)
				*activity.Longitude, _ = strconv.ParseFloat(longitude, 64)
			}
			activities = append(activities, activity)
		} else {
			log.Printf("rows.Scan failed: %v", err)
		}
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

func VerifyCredentials(username, password string) (user_id int64, authenticated bool) {
	row := db.QueryRow("SELECT id, pwhash, salt FROM users WHERE login = ? LIMIT 1", username)
	var db_hash []byte
	var salt []byte

	if err := row.Scan(&user_id, &db_hash, &salt); err != nil {
		log.Printf("VerifyCredentials: %v", err)
		return 0, false
	}

	password_hash := pbkdf2.Key([]byte(password), salt, PBKDF2_ROUNDS, PBKDF2_SIZE, sha256.New)

	return user_id, bytes.Equal(password_hash, db_hash)
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

	_, err = db.Exec("INSERT INTO users (login, pwhash, salt) VALUES (?, ?, ?)", username, password_hash, salt)
	if err != nil {
		log.Printf("RegisterUser: %v", err)
	}

	return err
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "can't anything other than POST", http.StatusMethodNotAllowed)
		return
	}

	r.Body.Close()

	session, _ := store.Get(r, SESSION_NAME)
	delete(session.Values, "Authenticated")
	delete(session.Values, "UserName")
	delete(session.Values, "UserId")
	session.Save(r, w)

	fmt.Fprintf(w, "OK")
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
	if user_id, ok := VerifyCredentials(username, password); ok {
		result.Authenticated = true

		rows, err := db.Query("SELECT id, name FROM activity_types WHERE user_id = ? AND active = 1", user_id)
		if err != nil {
			log.Printf("db.Query failed: %v", err)
		}

		result.Activities = []ActivityType{}
		for rows.Next() {
			var type_id int64
			var name string
			if err = rows.Scan(&type_id, &name); err == nil {
				result.Activities = append(result.Activities, ActivityType{Id: type_id, Name: name})
			}
		}

		// create new session and store that authentication was successful
		session, _ := store.Get(r, SESSION_NAME)
		session.Values["Authenticated"] = true
		session.Values["UserName"] = username
		session.Values["UserId"] = user_id
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
	var cfgfile *string = flag.String("config", "", "configuration file")
	flag.Parse()

	if *cfgfile == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := goconf.ReadConfigFile(*cfgfile)

	// TODO: add error handling
	driver, _ := cfg.GetString("database", "driver")
	dsn, _ := cfg.GetString("database", "dsn")
	auth_key, _ := cfg.GetString("sessions", "authkey")
	enc_key, _ := cfg.GetString("sessions", "enckey")

	db, err = sql.Open(driver, dsn)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	store = sessions.NewCookieStore([]byte(auth_key), []byte(enc_key))

	servemux := http.NewServeMux()

	servemux.Handle("/", http.FileServer(http.Dir("htdocs")))
	servemux.Handle("/auth", http.HandlerFunc(Authenticate))
	servemux.Handle("/auth/signup", http.HandlerFunc(Signup))
	servemux.Handle("/auth/logout", http.HandlerFunc(Logout))
	servemux.Handle("/activity/add", http.HandlerFunc(AddActivity))
	servemux.Handle("/activity/latest", http.HandlerFunc(LatestActivities))
	servemux.Handle("/activity/type/add", http.HandlerFunc(AddActivityType))

	httpsrv := &http.Server{Handler: servemux, Addr: ":8000"}
	if err := httpsrv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
