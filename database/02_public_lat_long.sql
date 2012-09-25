ALTER TABLE activities ADD COLUMN public INTEGER(1) NOT NULL;
UPDATE activities SET public = 0;

ALTER TABLE activities ADD COLUMN latitude DOUBLE;
ALTER TABLE activities ADD COLUMN longitude DOUBLE;
