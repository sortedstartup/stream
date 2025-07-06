-- Add channel support to videos
-- Videos can optionally belong to channels within tenants

-- Add channel_id field to videos table
ALTER TABLE videos ADD COLUMN channel_id TEXT;

-- Add indexes for performance
CREATE INDEX idx_videos_channel ON videos(channel_id);
CREATE INDEX idx_videos_tenant_channel ON videos(tenant_id, channel_id);

-- Note: channel_id references userservice_channels(id) but we don't add
-- a foreign key constraint since it's across databases 