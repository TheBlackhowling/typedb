-- Remove updated_at column from posts table
IF EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID(N'[dbo].[posts]') AND name = 'updated_at')
ALTER TABLE posts DROP COLUMN updated_at;

-- Remove updated_at column from users table
IF EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID(N'[dbo].[users]') AND name = 'updated_at')
ALTER TABLE users DROP COLUMN updated_at;
