-- Insert sample users
SET IDENTITY_INSERT users ON;
INSERT INTO users (id, name, email) VALUES
    (1, 'Alice', 'alice@example.com'),
    (2, 'Bob', 'bob@example.com'),
    (3, 'Charlie', 'charlie@example.com');
SET IDENTITY_INSERT users OFF;

-- Insert sample posts
SET IDENTITY_INSERT posts ON;
INSERT INTO posts (id, user_id, title, content, tags, metadata) VALUES
    (1, 1, 'First Post', 'This is my first post', '["go", "database"]', '{"views": 10, "likes": 5}'),
    (2, 1, 'Second Post', 'Another post', '["sql", "mssql"]', '{"views": 20, "likes": 8}'),
    (3, 2, 'Hello World', 'Hello from Bob', '["hello"]', '{"views": 5, "likes": 2}');
SET IDENTITY_INSERT posts OFF;

-- Insert sample user_posts relationships
INSERT INTO user_posts (user_id, post_id) VALUES
    (1, 1),
    (1, 2),
    (2, 3);
