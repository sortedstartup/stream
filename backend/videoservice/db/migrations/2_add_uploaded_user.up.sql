-- added uploaded_user column to videos table
ALTER TABLE videos ADD COLUMN uploaded_user_id TEXT NOT NULL;

-- added updated_at columns to videos table
ALTER TABLE videos ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;