-- Add service prefixes to commentservice tables
-- This migration renames tables to have consistent "commentservice_" prefix

-- Rename comments table to commentservice_comments
ALTER TABLE comments RENAME TO commentservice_comments;

-- Rename comment_likes table to commentservice_comment_likes
ALTER TABLE comment_likes RENAME TO commentservice_comment_likes;

-- Update foreign key references in commentservice_comment_likes
-- SQLite doesn't support ALTER CONSTRAINT directly, so we need to recreate the table

-- First, create new table with correct references
CREATE TABLE commentservice_comment_likes_new (
    id TEXT PRIMARY KEY,
    comment_id TEXT NOT NULL REFERENCES commentservice_comments(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(comment_id, user_id)
);

-- Copy data from old table (if any exists)
INSERT INTO commentservice_comment_likes_new (id, comment_id, user_id, created_at)
SELECT id, comment_id, user_id, created_at
FROM commentservice_comment_likes;

-- Drop old table
DROP TABLE commentservice_comment_likes;

-- Rename new table
ALTER TABLE commentservice_comment_likes_new RENAME TO commentservice_comment_likes;

-- Recreate indexes with new table names (if they existed)
-- Note: Adding indexes that should exist for performance
CREATE INDEX idx_commentservice_comments_video_id ON commentservice_comments(video_id);
CREATE INDEX idx_commentservice_comments_user_id ON commentservice_comments(user_id);
CREATE INDEX idx_commentservice_comments_created_at ON commentservice_comments(created_at);

CREATE INDEX idx_commentservice_comment_likes_comment_id ON commentservice_comment_likes(comment_id);
CREATE INDEX idx_commentservice_comment_likes_user_id ON commentservice_comment_likes(user_id); 