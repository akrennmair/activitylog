package main

import (
	"database/sql"
	"log"
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
