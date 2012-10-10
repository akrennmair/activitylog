-- create a table that contains all the flags including a name, that is a textual ID that will be mapped in the application
CREATE TABLE IF NOT EXISTS flags (
	id INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
	name VARCHAR(32) NOT NULL,
	UNIQUE(name)
) ENGINE=InnoDB;

-- insert flags
INSERT INTO flags (name) VALUES ('point_in_time'); -- activity type is a point in time (i.e. timestamp and end_timestamp are equal)
INSERT INTO flags (name) VALUES ('time_period');   -- activity type is a time period (i.e. end_timestamp is greater than timestamp)
INSERT INTO flags (name) VALUES ('numeric');       -- activity contains a numeric value in addition to the description
INSERT INTO flags (name) VALUES ('numeric_rateable'); -- activity contains a numeric value in addition to the description - shown as rating (e.g. 1 to 5 stars)

-- create a table that connects activity types and flags
CREATE TABLE IF NOT EXISTS activity_type_flags (
	id INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
	type_id INTEGER NOT NULL,
	flag_id INTEGER NOT NULL,
	FOREIGN KEY(type_id) REFERENCES activity_types(id),
	FOREIGN KEY(flag_id) REFERENCES flags(id)
) ENGINE=InnoDB;

-- add an end_timestamp to the activities table and set it to the timestamp
ALTER TABLE activities ADD COLUMN end_timestamp DATETIME;
UPDATE activities SET end_timestamp = timestamp WHERE timestamp IS NULL;

-- create activity_type_flags entries for all existing activity types
INSERT INTO activity_type_flags (type_id, flag_id) SELECT id, (SELECT id FROM flags WHERE name = 'point_in_time') FROM activities;
