-- MySQL Schema for typedb examples
-- Run this to set up the example database

CREATE DATABASE IF NOT EXISTS typedb_examples;
USE typedb_examples;

CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS posts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    tags JSON, -- MySQL JSON type
    metadata JSON, -- MySQL JSON type
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_posts (
    user_id INT NOT NULL,
    post_id INT NOT NULL,
    favorited_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

-- Insert sample data
INSERT INTO users (name, email) VALUES
    ('Alice', 'alice@example.com'),
    ('Bob', 'bob@example.com'),
    ('Charlie', 'charlie@example.com')
ON DUPLICATE KEY UPDATE name=name;

INSERT INTO posts (user_id, title, content, tags, metadata) VALUES
    (1, 'First Post', 'This is my first post', '["go", "database"]', '{"views": 10, "likes": 5}'),
    (1, 'Second Post', 'Another post', '["sql", "mysql"]', '{"views": 20, "likes": 8}'),
    (2, 'Hello World', 'Hello from Bob', '["hello"]', '{"views": 5, "likes": 2}')
ON DUPLICATE KEY UPDATE title=title;

INSERT INTO user_posts (user_id, post_id) VALUES
    (1, 1),
    (1, 2),
    (2, 3)
ON DUPLICATE KEY UPDATE user_id=user_id;
