-- Add service prefixes to videoservice tables
-- This migration renames tables to have consistent "videoservice_" prefix

-- Rename videos table to videoservice_videos
ALTER TABLE videos RENAME TO videoservice_videos;

-- Rename channels table to videoservice_channels
ALTER TABLE channels RENAME TO videoservice_channels;

-- Rename channel_members table to videoservice_channel_members
ALTER TABLE channel_members RENAME TO videoservice_channel_members;

-- Update foreign key references in videoservice_channel_members
-- SQLite doesn't support ALTER CONSTRAINT directly, so we need to recreate the table

-- First, create new table with correct references
CREATE TABLE videoservice_channel_members_new (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL REFERENCES videoservice_channels(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'viewer',
    added_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(channel_id, user_id)
);

-- Copy data from old table
INSERT INTO videoservice_channel_members_new (id, channel_id, user_id, role, added_by, created_at)
SELECT id, channel_id, user_id, role, added_by, created_at
FROM videoservice_channel_members;

-- Drop old table
DROP TABLE videoservice_channel_members;

-- Rename new table
ALTER TABLE videoservice_channel_members_new RENAME TO videoservice_channel_members;

-- Update foreign key references in videoservice_videos
-- SQLite doesn't support ALTER CONSTRAINT directly, so we need to recreate the table

-- First, create new table with correct references
CREATE TABLE videoservice_videos_new (
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
    FOREIGN KEY (channel_id) REFERENCES videoservice_channels(id) ON DELETE SET NULL
);

-- Copy data from old table
INSERT INTO videoservice_videos_new (id, title, description, url, created_at, uploaded_user_id, updated_at, is_private, tenant_id, channel_id)
SELECT id, title, description, url, created_at, uploaded_user_id, updated_at, is_private, tenant_id, channel_id
FROM videoservice_videos;

-- Drop old table
DROP TABLE videoservice_videos;

-- Rename new table
ALTER TABLE videoservice_videos_new RENAME TO videoservice_videos;

-- Recreate all indexes with new table names
CREATE INDEX idx_videoservice_videos_tenant_id ON videoservice_videos(tenant_id);
CREATE INDEX idx_videoservice_videos_tenant_user ON videoservice_videos(tenant_id, uploaded_user_id);
CREATE INDEX idx_videoservice_videos_channel ON videoservice_videos(channel_id);
CREATE INDEX idx_videoservice_videos_tenant_channel ON videoservice_videos(tenant_id, channel_id);

CREATE INDEX idx_videoservice_channels_tenant ON videoservice_channels(tenant_id);
CREATE INDEX idx_videoservice_channels_created_by ON videoservice_channels(created_by);
CREATE INDEX idx_videoservice_channels_tenant_created ON videoservice_channels(tenant_id, created_at);

CREATE INDEX idx_videoservice_channel_members_channel ON videoservice_channel_members(channel_id);
CREATE INDEX idx_videoservice_channel_members_user ON videoservice_channel_members(user_id);
CREATE INDEX idx_videoservice_channel_members_role ON videoservice_channel_members(role); 