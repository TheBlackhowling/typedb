-- SQLite Schema for typedb examples
-- Run this to set up the example database

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT,
    tags TEXT, -- JSON stored as TEXT in SQLite
    metadata TEXT, -- JSON stored as TEXT in SQLite
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_posts (
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    favorited_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

-- Insert sample data
INSERT OR IGNORE INTO users (id, name, email) VALUES
    (1, 'Alice', 'alice@example.com'),
    (2, 'Bob', 'bob@example.com'),
    (3, 'Charlie', 'charlie@example.com');

INSERT OR IGNORE INTO posts (id, user_id, title, content, tags, metadata) VALUES
    (1, 1, 'First Post', 'This is my first post', '["go", "database"]', '{"views": 10, "likes": 5}'),
    (2, 1, 'Second Post', 'Another post', '["sql", "sqlite"]', '{"views": 20, "likes": 8}'),
    (3, 2, 'Hello World', 'Hello from Bob', '["hello"]', '{"views": 5, "likes": 2}');

INSERT OR IGNORE INTO user_posts (user_id, post_id) VALUES
    (1, 1),
    (1, 2),
    (2, 3);
