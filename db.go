package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"log"
	"strconv"
	"code.google.com/p/go.crypto/pbkdf2"
	"crypto/sha256"
)

type ActivityTypesGetter interface {
	GetActivityTypesForUser(user_id int64) []ActivityType
}

type Database struct {
	conn *sql.DB
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

func GetActivitiesForUser(user_id int64, limit uint, offset uint) ([]Activity, error) {
	rows, err := db.Query("SELECT type_id, timestamp, description, latitude, longitude FROM activities WHERE user_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?", user_id, limit, offset)
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
