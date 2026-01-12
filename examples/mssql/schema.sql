-- Simplified schema for examples (happy path demonstrations)
-- Use migrations for actual database setup

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(255) NOT NULL,
    email NVARCHAR(255) UNIQUE NOT NULL,
    created_at DATETIME2 DEFAULT GETDATE()
);

-- Create profiles table (one-to-one with users)
CREATE TABLE IF NOT EXISTS profiles (
    id INT IDENTITY(1,1) PRIMARY KEY,
    user_id INT NOT NULL UNIQUE,
    bio NVARCHAR(MAX),
    avatar_url NVARCHAR(500),
    location NVARCHAR(255),
    website NVARCHAR(500),
    created_at DATETIME2 DEFAULT GETDATE(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create posts table (many-to-one with users)
CREATE TABLE IF NOT EXISTS posts (
    id INT IDENTITY(1,1) PRIMARY KEY,
    user_id INT NOT NULL,
    title NVARCHAR(255) NOT NULL,
    content NVARCHAR(MAX),
    published BIT DEFAULT 0,
    created_at DATETIME2 DEFAULT GETDATE(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
