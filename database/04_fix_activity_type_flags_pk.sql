ALTER TABLE activity_type_flags MODIFY id INTEGER DEFAULT NULL;
ALTER TABLE activity_type_flags DROP PRIMARY KEY, ADD PRIMARY KEY (type_id, flag_id);
ALTER TABLE activity_type_flags DROP COLUMN id;
