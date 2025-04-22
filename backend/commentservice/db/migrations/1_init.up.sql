PRAGMA foreign_keys = ON;

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE comments (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    video_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    parent_comment_id TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_video_id ON comments(video_id);
CREATE INDEX idx_user_id ON comments(user_id);
CREATE INDEX idx_parent_comment_id ON comments(parent_comment_id);

CREATE TABLE comment_likes (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    comment_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
