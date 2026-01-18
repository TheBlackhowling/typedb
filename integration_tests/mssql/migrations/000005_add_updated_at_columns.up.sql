-- Add updated_at column to users table
IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID(N'[dbo].[users]') AND name = 'updated_at')
ALTER TABLE users ADD updated_at DATETIME2;

-- Add updated_at column to posts table
IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID(N'[dbo].[posts]') AND name = 'updated_at')
ALTER TABLE posts ADD updated_at DATETIME2;
