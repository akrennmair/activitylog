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
	AddActivityType(typename string, user_id int64) (ActivityType, error)
}

type ActivityTypeUpdater interface {
	UpdateActivityType(typename string, user_id int64, activity_type_id int64) error
}

type ActivityTypeDeleter interface {
	DeleteActivityType(user_id int64, activity_type_id int64) error
}

type UserRegistrar interface {
	RegisterUser(username, password string) error
}

type Database struct {
	conn *sql.DB
}

func (db *Database) UpdateActivityType(typename string, user_id int64, activity_type_id int64) error {
	_, err := db.conn.Exec("UPDATE activity_types SET name = ? WHERE user_id = ? AND id = ?", typename, user_id, activity_type_id)
	if err != nil {
		log.Printf("db.conn.Exec failed: %v", err)
	}
	return err
}

func (db *Database) AddActivityType(typename string, user_id int64) (activity_type ActivityType, err error) {
	activity_type = ActivityType{Name: typename}
	result, err := db.conn.Exec("INSERT INTO activity_types (name, user_id, active) VALUES (?, ?, 1)", typename, user_id)
	if err != nil {
		log.Printf("db.conn.Exec failed: %v", err)
		return activity_type, err
	}

	activity_type.Id, _ = result.LastInsertId()
	return activity_type, nil
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

	_, err = db.conn.Exec("INSERT INTO activities (type_id, timestamp, description, user_id, public, latitude, longitude) VALUES (?, NOW(), ?, ?, ?, ?, ?)", type_id, description, user_id, public, latitude, longitude)
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
			activities = append(activities, ActivityType{Id: type_id, Name: name})
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
