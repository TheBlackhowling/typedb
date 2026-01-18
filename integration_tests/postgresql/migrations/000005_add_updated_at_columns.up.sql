-- Add updated_at column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP;

-- Add updated_at column to posts table
ALTER TABLE posts ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP;
