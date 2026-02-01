-- Remove phone column from users table
IF EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID(N'[dbo].[users]') AND name = 'phone')
BEGIN
    ALTER TABLE users DROP COLUMN phone;
END
