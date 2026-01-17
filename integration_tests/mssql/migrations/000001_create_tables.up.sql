-- Create users table
CREATE TABLE users (
    id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(255) NOT NULL,
    email NVARCHAR(255) UNIQUE NOT NULL,
    created_at DATETIME2 DEFAULT GETDATE()
);

-- Create posts table
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

-- Create user_posts junction table
-- Note: Using NO ACTION on user_id FK to avoid multiple cascade paths
-- (posts already cascades from users, so user_posts->users would create a cycle)
CREATE TABLE user_posts (
    user_id INT NOT NULL,
    post_id INT NOT NULL,
    favorited_at DATETIME2 DEFAULT GETDATE(),
    PRIMARY KEY (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE NO ACTION,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);
