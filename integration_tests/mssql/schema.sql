-- SQL Server Schema for typedb examples
-- Run this to set up the example database

USE master;
GO

IF NOT EXISTS (SELECT * FROM sys.databases WHERE name = 'typedb_examples')
BEGIN
    CREATE DATABASE typedb_examples;
END
GO

USE typedb_examples;
GO

IF OBJECT_ID('user_posts', 'U') IS NOT NULL DROP TABLE user_posts;
IF OBJECT_ID('posts', 'U') IS NOT NULL DROP TABLE posts;
IF OBJECT_ID('users', 'U') IS NOT NULL DROP TABLE users;
GO

CREATE TABLE users (
    id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(255) NOT NULL,
    email NVARCHAR(255) UNIQUE NOT NULL,
    created_at DATETIME2 DEFAULT GETDATE()
);

CREATE TABLE posts (
    id INT IDENTITY(1,1) PRIMARY KEY,
    user_id INT NOT NULL,
    title NVARCHAR(255) NOT NULL,
    content NVARCHAR(MAX),
    tags NVARCHAR(MAX), -- JSON stored as NVARCHAR(MAX)
    metadata NVARCHAR(MAX), -- JSON stored as NVARCHAR(MAX)
    created_at DATETIME2 DEFAULT GETDATE(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE user_posts (
    user_id INT NOT NULL,
    post_id INT NOT NULL,
    favorited_at DATETIME2 DEFAULT GETDATE(),
    PRIMARY KEY (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

-- Insert sample data
INSERT INTO users (name, email) VALUES
    ('Alice', 'alice@example.com'),
    ('Bob', 'bob@example.com'),
    ('Charlie', 'charlie@example.com');

INSERT INTO posts (user_id, title, content, tags, metadata) VALUES
    (1, 'First Post', 'This is my first post', '["go", "database"]', '{"views": 10, "likes": 5}'),
    (1, 'Second Post', 'Another post', '["sql", "mssql"]', '{"views": 20, "likes": 8}'),
    (2, 'Hello World', 'Hello from Bob', '["hello"]', '{"views": 5, "likes": 2}');

INSERT INTO user_posts (user_id, post_id) VALUES
    (1, 1),
    (1, 2),
    (2, 3);
GO
