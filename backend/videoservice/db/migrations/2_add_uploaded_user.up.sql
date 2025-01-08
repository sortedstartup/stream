-- added uploaded_user column to videos table
ALTER TABLE videos ADD COLUMN uploaded_user_id TEXT NOT NULL;