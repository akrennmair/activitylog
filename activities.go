package main

import (
	"bytes"
	_ "code.google.com/p/go-mysql-driver/mysql"
	"code.google.com/p/go.crypto/pbkdf2"
	"code.google.com/p/gorilla/pat"
	"code.google.com/p/gorilla/sessions"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"
	goconf "github.com/akrennmair/goconf"
	"log"
	"net/http"
	"os"
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
	PBKDF2_ROUNDS = 10000
	PBKDF2_SIZE   = 32
)

type ActivityType struct {
	Id   int64  `json:"type_id"`
	Name string `json:"name"`
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

	dbx := &Database{conn: db}

	store = sessions.NewCookieStore([]byte(auth_key), []byte(enc_key))

	r := pat.New()

	r.Add("POST", "/auth/try", &TryAuthenticateHandler{Db: dbx})
	r.Add("POST", "/auth/signup", &SignupHandler{})
	r.Post("/auth/logout", http.HandlerFunc(Logout))
	r.Add("POST", "/auth", &AuthenticateHandler{Db: dbx})
	r.Add("POST", "/activity/add", &AddActivityHandler{})
	r.Add("GET", "/activity/list/{page:[0-9]+}", &ListActivitiesHandler{})
	r.Add("POST", "/activity/type/add", &AddActivityTypeHandler{})
	r.Add("POST", "/activity/type/edit", &EditActivityTypeHandler{})
	r.Add("POST", "/activity/type/del", &DeleteActivityTypeHandler{})
	r.Add("GET", "/activity/type/list", &ListActivityTypesHandler{Db: dbx})

	r.Add("GET", "/activity/latest", &LatestActivitiesHandler{})
	r.Add("GET", "/", http.FileServer(http.Dir("htdocs")))

	httpsrv := &http.Server{Handler: r, Addr: ":8000"}
	if err := httpsrv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
