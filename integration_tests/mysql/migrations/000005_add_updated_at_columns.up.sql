-- Add updated_at column to users table (TIMESTAMP(6) for microsecond precision)
ALTER TABLE users ADD COLUMN updated_at TIMESTAMP(6);

-- Add updated_at column to posts table (TIMESTAMP(6) for microsecond precision)
ALTER TABLE posts ADD COLUMN updated_at TIMESTAMP(6);
