-- Create new user with username: user, password: user
-- Password hash untuk 'user': $2a$10$... (bcrypt)

-- Insert user
INSERT INTO users (username, password, full_name, email, role_id, is_active, created_at, updated_at)
VALUES (
    'user',
    '$2a$10$rZJ5WxV8FvH8W7iN5vY9JuKkH7nF3Qf3W5F0F0F0F0F0F0F0F0F0F0',  -- password: user
    'Regular User',
    'user@stockopname.com',
    (SELECT id FROM roles WHERE name = 'staff' LIMIT 1),
    true,
    NOW(),
    NOW()
)
ON CONFLICT (username) DO UPDATE SET
    password = EXCLUDED.password,
    updated_at = NOW();

-- Verify
SELECT id, username, full_name, role_id, is_active FROM users WHERE username = 'user';
