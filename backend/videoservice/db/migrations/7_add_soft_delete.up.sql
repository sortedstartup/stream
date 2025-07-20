-- Add soft delete functionality to videos
-- Videos can be soft deleted to hide them from all views while preserving data

ALTER TABLE videoservice_videos ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;

-- Add index for performance when filtering out deleted videos
CREATE INDEX idx_videoservice_videos_is_deleted ON videoservice_videos(is_deleted);

-- Compound index for common queries (tenant + not deleted)
CREATE INDEX idx_videoservice_videos_tenant_not_deleted ON videoservice_videos(tenant_id, is_deleted); 