-- Remove foreign key constraint by recreating comments table without it
-- Keep the users table since it's now managed by user service
PRAGMA foreign_keys = OFF;

-- Create temporary table without foreign key constraint
CREATE TABLE comments_temp (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    video_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    parent_comment_id TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    username TEXT
);

-- Copy data from original table
INSERT INTO comments_temp 
SELECT id, content, video_id, user_id, parent_comment_id, created_at, updated_at, username 
FROM comments;

-- Drop original table
DROP TABLE comments;

-- Rename temp table to original name
ALTER TABLE comments_temp RENAME TO comments;

-- Recreate indexes
CREATE INDEX idx_video_id ON comments(video_id);
CREATE INDEX idx_user_id ON comments(user_id);
CREATE INDEX idx_parent_comment_id ON comments(parent_comment_id);

-- Note: We keep the users table since it's now managed by user service

PRAGMA foreign_keys = ON; 