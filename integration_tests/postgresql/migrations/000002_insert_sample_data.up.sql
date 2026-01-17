-- Insert sample users
INSERT INTO users (name, email) VALUES
    ('Alice', 'alice@example.com'),
    ('Bob', 'bob@example.com'),
    ('Charlie', 'charlie@example.com')
ON CONFLICT (email) DO NOTHING;

-- Insert sample posts
INSERT INTO posts (user_id, title, content, tags, metadata) VALUES
    (1, 'First Post', 'This is my first post', ARRAY['go', 'database'], '{"views": 10, "likes": 5}'),
    (1, 'Second Post', 'Another post', ARRAY['sql', 'postgres'], '{"views": 20, "likes": 8}'),
    (2, 'Hello World', 'Hello from Bob', ARRAY['hello'], '{"views": 5, "likes": 2}')
ON CONFLICT DO NOTHING;

-- Insert sample user_posts relationships
INSERT INTO user_posts (user_id, post_id) VALUES
    (1, 1),
    (1, 2),
    (2, 3)
ON CONFLICT DO NOTHING;
