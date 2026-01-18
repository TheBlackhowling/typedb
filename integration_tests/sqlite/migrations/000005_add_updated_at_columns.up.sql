-- Add updated_at column to users table
ALTER TABLE users ADD COLUMN updated_at DATETIME;

-- Add updated_at column to posts table
ALTER TABLE posts ADD COLUMN updated_at DATETIME;
