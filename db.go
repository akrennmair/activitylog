package main

import (
	"bytes"
	"code.google.com/p/go.crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"log"
	"strconv"
)

type ActivityTypesGetter interface {
	GetActivityTypesForUser(user_id int64) []ActivityType
}

type CredentialsVerifierActivityTypesGetter interface {
	ActivityTypesGetter
	VerifyCredentials(username, password string) (user_id int64, authenticated bool)
}

type ActivityAdder interface {
	AddActivity(type_id int64, description string, user_id int64, is_public bool, latitude, longitude string) error
}

type ActivityTypeAdder interface {
	AddActivityType(typename string, user_id int64, time_period bool) (ActivityType, error)
}

type ActivityTypeRenamer interface {
	RenameActivityType(typename string, user_id int64, activity_type_id int64) error
}

type ActivityTypeDeleter interface {
	DeleteActivityType(user_id int64, activity_type_id int64) error
}

type UserRegistrar interface {
	RegisterUser(username, password string) error
}

type Database struct {
	conn *sql.DB
	flagNameToIdMapping map[string]int64
	flagIdToNameMapping map[int64]string
}

const (
	FLAG_TIME_PERIOD = "time_period"
	FLAG_POINT_IN_TIME = "point_in_time"
)

func NewDatabase(conn *sql.DB) *Database {
	db := &Database{conn: conn}
	db.flagNameToIdMapping = make(map[string]int64)
	db.flagIdToNameMapping = make(map[int64]string)
	rows, err := db.conn.Query("SELECT id, name FROM flags")
	if err != nil {
		log.Fatalf("NewDatabase: loading flag names failed: %v", err)
	}
	for rows.Next() {
		var id int64
		var name string
		if err = rows.Scan(&id, &name); err == nil {
			db.flagNameToIdMapping[name] = id
			db.flagIdToNameMapping[id] = name
		} else {
			log.Printf("NewDatabase: loading flag name in rows.Scan failed: %v", err)
		}
	}
	return db
}

func (db *Database) RenameActivityType(typename string, user_id int64, activity_type_id int64) error {
	_, err := db.conn.Exec("UPDATE activity_types SET name = ? WHERE user_id = ? AND id = ?", typename, user_id, activity_type_id)
	if err != nil {
		log.Printf("db.conn.Exec failed: %v", err)
	}
	return err
}

func (db *Database) AddActivityType(typename string, user_id int64, time_period bool) (activity_type ActivityType, err error) {
	activity_type = ActivityType{Name: typename}

	txn, err := db.conn.Begin()
	if err != nil {
		log.Printf("starting transaction failed: %v")
		return activity_type, err
	}

	result, err := txn.Exec("INSERT INTO activity_types (name, user_id, active) VALUES (?, ?, 1)", typename, user_id)
	if err != nil {
		log.Printf("db.conn.Exec failed: %v", err)
		txn.Rollback()
		return activity_type, err
	}

	activity_type.Id, _ = result.LastInsertId()

	// set flag time_period/point_in_time depending on time_period argument
	flag_id := db.flagNameToIdMapping[FLAG_POINT_IN_TIME]
	if time_period {
		flag_id = db.flagNameToIdMapping[FLAG_TIME_PERIOD]
	}

	result, err = txn.Exec("INSERT INTO activity_type_flags (type_id, flag_id) VALUES (?, ?)", activity_type.Id, flag_id)
	if err != nil {
		log.Printf("inserting flag failed: %v", err)
		txn.Rollback()
		return activity_type, err
	}

	activity_type.TimePeriod = time_period

	err = txn.Commit()
	if err != nil {
		log.Printf("committing transaction failed: %v", err)
	}

	return activity_type, err
}

func (db *Database) AddActivity(type_id int64, description string, user_id int64, is_public bool, lat, long string) error {
	var latitude sql.NullFloat64
	var longitude sql.NullFloat64

	var err error
	if latitude.Float64, err = strconv.ParseFloat(lat, 64); err != nil {
		latitude.Valid = false
	} else {
		latitude.Valid = true
	}

	if longitude.Float64, err = strconv.ParseFloat(long, 64); err != nil {
		longitude.Valid = false
	} else {
		longitude.Valid = true
	}

	public := 0
	if is_public {
		public = 1
	}

	txn, err := db.conn.Begin()
	if err != nil {
		log.Printf("Starting transaction failed: %v", err)
		return err
	}

	row := txn.QueryRow("SELECT count(1) FROM activity_type_flags WHERE type_id = ? AND flag_id = ?", type_id, db.flagNameToIdMapping[FLAG_POINT_IN_TIME])
	var point_in_time_count int64
	if err = row.Scan(&point_in_time_count); err != nil {
		log.Printf("row.Scan failed: %v", err)
		txn.Rollback()
		return err
	}

	result, err := txn.Exec("INSERT INTO activities (type_id, timestamp, description, user_id, public, latitude, longitude) VALUES (?, NOW(), ?, ?, ?, ?, ?)", type_id, description, user_id, public, latitude, longitude)
	if err == nil {
		activity_id, _ := result.LastInsertId()
		if point_in_time_count != 0 {
			_, err = txn.Exec("UPDATE activities SET end_timestamp = NOW() WHERE id = ?", activity_id)
			if err != nil {
				log.Printf("db.conn.Exec failed: %v", err)
			}
		}
	}

	if err != nil {
		txn.Rollback()
	} else {
		err = txn.Commit()
		if err != nil {
			log.Printf("committing transaction failed: %v", err)
		}
	}
	return err
}

func (db *Database) GetActivityTypesForUser(user_id int64) (activities []ActivityType) {
	activities = []ActivityType{}

	rows, err := db.conn.Query("SELECT id, name FROM activity_types WHERE user_id = ? AND active = 1", user_id)
	if err != nil {
		log.Printf("db.conn.Query failed: %v", err)
		return
	}

	for rows.Next() {
		var type_id int64
		var name string
		if err = rows.Scan(&type_id, &name); err == nil {
			time_period := false
			if rows_flags, err := db.conn.Query("SELECT flag_id FROM activity_type_flags WHERE type_id = ?", type_id); err == nil {
				for rows_flags.Next() {
					var flag_id int64
					if err = rows.Scan(&flag_id); err == nil {
						switch flag_id {
						case db.flagNameToIdMapping[FLAG_TIME_PERIOD]:
							time_period = true
						case db.flagNameToIdMapping[FLAG_POINT_IN_TIME]:
							time_period = false
						}
					}
				}
			}
			activities = append(activities, ActivityType{Id: type_id, Name: name, TimePeriod: time_period})
		}
	}
	return
}

func (db *Database) GetActivitiesForUser(user_id int64, limit uint, offset uint) ([]Activity, error) {
	rows, err := db.conn.Query("SELECT type_id, timestamp, description, latitude, longitude FROM activities WHERE user_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?", user_id, limit, offset)
	if err != nil {
		log.Printf("GetActivitiesForUser: %v", err)
		return nil, err
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

	return activities, nil
}

func (db *Database) RegisterUser(username, password string) error {
	salt, err := GenerateSalt()
	if err != nil {
		return err
	}

	password_hash := HashPassword([]byte(password), salt)

	_, err = db.conn.Exec("INSERT INTO users (login, pwhash, salt) VALUES (?, ?, ?)", username, password_hash, salt)
	if err != nil {
		log.Printf("RegisterUser: %v", err)
	}

	return err
}

func (db *Database) VerifyCredentials(username, password string) (user_id int64, authenticated bool) {
	row := db.conn.QueryRow("SELECT id, pwhash, salt FROM users WHERE login = ? LIMIT 1", username)
	var db_hash []byte
	var salt []byte

	if err := row.Scan(&user_id, &db_hash, &salt); err != nil {
		log.Printf("VerifyCredentials: %v", err)
		return 0, false
	}

	password_hash := HashPassword([]byte(password), salt)

	return user_id, bytes.Equal(password_hash, db_hash)
}

func (db *Database) DeleteActivityType(user_id int64, activity_type_id int64) error {
	_, err := db.conn.Exec("UPDATE activity_types SET active = 0 WHERE user_id = ? AND id = ?", user_id, activity_type_id)
	return err
}

func GenerateSalt() (data []byte, err error) {
	data = make([]byte, 8)
	_, err = rand.Read(data)
	return
}

func HashPassword(password, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, PBKDF2_ROUNDS, PBKDF2_SIZE, sha256.New)
}
