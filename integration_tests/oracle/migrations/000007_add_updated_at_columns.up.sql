-- Add updated_at column to users table
ALTER TABLE users ADD updated_at TIMESTAMP;

-- Add updated_at column to posts table
ALTER TABLE posts ADD updated_at TIMESTAMP;
