-- Remove updated_at column from posts table
ALTER TABLE posts DROP COLUMN updated_at;

-- Remove updated_at column from users table
ALTER TABLE users DROP COLUMN updated_at;
