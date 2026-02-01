-- Add phone column to users table
IF NOT EXISTS (SELECT * FROM sys.columns WHERE object_id = OBJECT_ID(N'[dbo].[users]') AND name = 'phone')
BEGIN
    ALTER TABLE users ADD phone NVARCHAR(20);
END
