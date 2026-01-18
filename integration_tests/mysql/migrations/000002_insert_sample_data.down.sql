-- Remove sample data
DELETE FROM user_posts WHERE user_id IN (1, 2) AND post_id IN (1, 2, 3);
DELETE FROM posts WHERE id IN (1, 2, 3);
DELETE FROM users WHERE id IN (1, 2, 3);
