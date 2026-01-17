-- Insert sample users
INSERT OR IGNORE INTO users (id, name, email) VALUES
    (1, 'Alice', 'alice@example.com'),
    (2, 'Bob', 'bob@example.com'),
    (3, 'Charlie', 'charlie@example.com');

-- Insert sample posts
INSERT OR IGNORE INTO posts (id, user_id, title, content, tags, metadata) VALUES
    (1, 1, 'First Post', 'This is my first post', '["go", "database"]', '{"views": 10, "likes": 5}'),
    (2, 1, 'Second Post', 'Another post', '["sql", "sqlite"]', '{"views": 20, "likes": 8}'),
    (3, 2, 'Hello World', 'Hello from Bob', '["hello"]', '{"views": 5, "likes": 2}');

-- Insert sample user_posts relationships
INSERT OR IGNORE INTO user_posts (user_id, post_id) VALUES
    (1, 1),
    (1, 2),
    (2, 3);
