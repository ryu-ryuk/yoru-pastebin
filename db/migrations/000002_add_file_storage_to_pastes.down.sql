DELETE FROM pastes WHERE is_file = TRUE;

ALTER TABLE pastes
DROP COLUMN is_file,
DROP COLUMN file_name,
DROP COLUMN mime_type,
DROP COLUMN file_size,
DROP COLUMN s3_key;

ALTER TABLE pastes ALTER COLUMN content SET NOT NULL;
