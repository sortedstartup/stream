-- Create spaces table
CREATE TABLE spaces (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    user_id TEXT NOT NULL, -- owner of the space
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create user_spaces table for space members
CREATE TABLE user_spaces (
    user_id TEXT NOT NULL,
    space_id TEXT NOT NULL,
    access_level TEXT NOT NULL DEFAULT 'view', -- view, edit, admin
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, space_id),
    FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE
);

-- Create video_spaces table for video assignments
CREATE TABLE video_spaces (
    video_id TEXT NOT NULL,
    space_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (video_id, space_id),
    FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE,
    FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE
);

-- Add indexes for better performance
CREATE INDEX idx_spaces_user_id ON spaces(user_id);
CREATE INDEX idx_user_spaces_user_id ON user_spaces(user_id);
CREATE INDEX idx_user_spaces_space_id ON user_spaces(space_id);
CREATE INDEX idx_video_spaces_video_id ON video_spaces(video_id);
CREATE INDEX idx_video_spaces_space_id ON video_spaces(space_id); 