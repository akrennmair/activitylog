package main

import (
	_ "code.google.com/p/go-mysql-driver/mysql"
	"code.google.com/p/gorilla/pat"
	"code.google.com/p/gorilla/sessions"
	"database/sql"
	"flag"
	goconf "github.com/akrennmair/goconf"
	"log"
	"net/http"
	"os"
)


const (
	ActivityLimit = 10
	SESSION_NAME  = "activities-session"
	PBKDF2_ROUNDS = 10000
	PBKDF2_SIZE   = 32
)


type Activity struct {
	TypeId      int64    `json:"type_id"`
	Timestamp   string   `json:"ts"`
	Description string   `json:"desc"`
	Latitude    *float64 `json:"lat"`
	Longitude   *float64 `json:"long"`
}


type ActivityType struct {
	Id   int64  `json:"type_id"`
	Name string `json:"name"`
}

type AuthResult struct {
	Authenticated bool           `json:"authenticated"`
	ErrorMsg      string         `json:"errormsg,omitempty"`
	Activities    []ActivityType `json:"activities,omitempty"`
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

	db_handle, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	defer db_handle.Close()

	db := &Database{conn: db_handle}

	store := sessions.NewCookieStore([]byte(auth_key), []byte(enc_key))

	r := pat.New()

	r.Add("POST", "/auth/try", &TryAuthenticateHandler{Db: db, Store: store})
	r.Add("POST", "/auth/signup", &SignupHandler{Db: db})
	r.Add("POST", "/auth/logout", &LogoutHandler{Store: store})
	r.Add("POST", "/auth", &AuthenticateHandler{Db: db, Store: store})
	r.Add("POST", "/activity/add", &AddActivityHandler{Store: store, Db: db})
	r.Add("GET",  "/activity/list/{page:[0-9]+}", &ListActivitiesHandler{Db: db, Store: store})
	r.Add("POST", "/activity/type/add", &AddActivityTypeHandler{Db: db, Store: store})
	r.Add("POST", "/activity/type/edit", &EditActivityTypeHandler{/* Db: db, */Store: store})
	r.Add("POST", "/activity/type/del", &DeleteActivityTypeHandler{Db: db, Store: store})
	r.Add("GET",  "/activity/type/list", &ListActivityTypesHandler{Db: db, Store: store})
	r.Add("GET",  "/activity/latest", &LatestActivitiesHandler{Db: db, Store: store})
	r.Add("GET",  "/", http.FileServer(http.Dir("htdocs")))

	httpsrv := &http.Server{Handler: r, Addr: ":8000"}
	if err := httpsrv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
