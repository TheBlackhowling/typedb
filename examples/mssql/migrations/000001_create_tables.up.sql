-- Create users table
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[users]') AND type in (N'U'))
BEGIN
    CREATE TABLE users (
        id INT IDENTITY(1,1) PRIMARY KEY,
        name NVARCHAR(255) NOT NULL,
        email NVARCHAR(255) UNIQUE NOT NULL,
        created_at DATETIME2 DEFAULT GETDATE()
    );
END

-- Create profiles table (one-to-one with users)
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[profiles]') AND type in (N'U'))
BEGIN
    CREATE TABLE profiles (
        id INT IDENTITY(1,1) PRIMARY KEY,
        user_id INT NOT NULL UNIQUE,
        bio NVARCHAR(MAX),
        avatar_url NVARCHAR(500),
        location NVARCHAR(255),
        website NVARCHAR(500),
        created_at DATETIME2 DEFAULT GETDATE(),
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );
END

-- Create posts table (many-to-one with users)
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[posts]') AND type in (N'U'))
BEGIN
    CREATE TABLE posts (
        id INT IDENTITY(1,1) PRIMARY KEY,
        user_id INT NOT NULL,
        title NVARCHAR(255) NOT NULL,
        content NVARCHAR(MAX),
        published BIT DEFAULT 0,
        created_at DATETIME2 DEFAULT GETDATE(),
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );
END

-- Create user_posts table (many-to-many relationship with composite primary key)
-- Note: SQL Server doesn't allow multiple CASCADE paths, so we use NO ACTION for post_id
IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[user_posts]') AND type in (N'U'))
BEGIN
    CREATE TABLE user_posts (
        user_id INT NOT NULL,
        post_id INT NOT NULL,
        favorited_at DATETIME2 DEFAULT GETDATE(),
        PRIMARY KEY (user_id, post_id),
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE NO ACTION
    );
END
