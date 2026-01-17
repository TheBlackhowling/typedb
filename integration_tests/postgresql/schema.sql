-- PostgreSQL Schema for typedb examples
-- Run this to set up the example database

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    tags TEXT[], -- PostgreSQL array
    metadata JSONB, -- PostgreSQL JSONB
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_posts (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    favorited_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, post_id)
);

-- Insert sample data
INSERT INTO users (name, email) VALUES
    ('Alice', 'alice@example.com'),
    ('Bob', 'bob@example.com'),
    ('Charlie', 'charlie@example.com')
ON CONFLICT (email) DO NOTHING;

INSERT INTO posts (user_id, title, content, tags, metadata) VALUES
    (1, 'First Post', 'This is my first post', ARRAY['go', 'database'], '{"views": 10, "likes": 5}'),
    (1, 'Second Post', 'Another post', ARRAY['sql', 'postgres'], '{"views": 20, "likes": 8}'),
    (2, 'Hello World', 'Hello from Bob', ARRAY['hello'], '{"views": 5, "likes": 2}')
ON CONFLICT DO NOTHING;

INSERT INTO user_posts (user_id, post_id) VALUES
    (1, 1),
    (1, 2),
    (2, 3)
ON CONFLICT DO NOTHING;
