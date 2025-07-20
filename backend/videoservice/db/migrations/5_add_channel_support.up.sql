-- Add channel support to videos
-- Videos can optionally belong to channels within tenants

-- SQLite doesn't support adding foreign key constraints with ALTER TABLE ADD COLUMN
-- So we need to recreate the table to add the constraint

-- Create new table with channel_id and foreign key constraint
CREATE TABLE videos_new (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    uploaded_user_id TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_private BOOLEAN DEFAULT TRUE,
    tenant_id TEXT,
    channel_id TEXT,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE SET NULL
);

-- Copy data from old table
INSERT INTO videos_new (id, title, description, url, created_at, uploaded_user_id, updated_at, is_private, tenant_id)
SELECT id, title, description, url, created_at, uploaded_user_id, updated_at, is_private, tenant_id
FROM videos;

-- Drop old table
DROP TABLE videos;

-- Rename new table
ALTER TABLE videos_new RENAME TO videos;

-- Recreate indexes
CREATE INDEX idx_videos_tenant_id ON videos(tenant_id);
CREATE INDEX idx_videos_tenant_user ON videos(tenant_id, uploaded_user_id);
CREATE INDEX idx_videos_channel ON videos(channel_id);
CREATE INDEX idx_videos_tenant_channel ON videos(tenant_id, channel_id);