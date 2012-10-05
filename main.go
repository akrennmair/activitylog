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

var (
	db    *sql.DB
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

	db, err = sql.Open(driver, dsn)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	dbx := &Database{conn: db}

	store := sessions.NewCookieStore([]byte(auth_key), []byte(enc_key))

	r := pat.New()

	r.Add("POST", "/auth/try", &TryAuthenticateHandler{Db: dbx, Store: store})
	r.Add("POST", "/auth/signup", &SignupHandler{})
	r.Add("POST", "/auth/logout", &LogoutHandler{Store: store})
	r.Add("POST", "/auth", &AuthenticateHandler{Db: dbx, Store: store})
	r.Add("POST", "/activity/add", &AddActivityHandler{Store: store})
	r.Add("GET",  "/activity/list/{page:[0-9]+}", &ListActivitiesHandler{Store: store})
	r.Add("POST", "/activity/type/add", &AddActivityTypeHandler{Store: store})
	r.Add("POST", "/activity/type/edit", &EditActivityTypeHandler{Store: store})
	r.Add("POST", "/activity/type/del", &DeleteActivityTypeHandler{Store: store})
	r.Add("GET",  "/activity/type/list", &ListActivityTypesHandler{Db: dbx, Store: store})
	r.Add("GET",  "/activity/latest", &LatestActivitiesHandler{Store: store})
	r.Add("GET",  "/", http.FileServer(http.Dir("htdocs")))

	httpsrv := &http.Server{Handler: r, Addr: ":8000"}
	if err := httpsrv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
