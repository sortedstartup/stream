-- Add privacy and tenant support to videos
ALTER TABLE videos ADD COLUMN is_private BOOLEAN DEFAULT TRUE;
ALTER TABLE videos ADD COLUMN tenant_id TEXT;

-- Add index for tenant-based queries
CREATE INDEX idx_videos_tenant_id ON videos(tenant_id);
CREATE INDEX idx_videos_tenant_user ON videos(tenant_id, uploaded_user_id);

-- Set existing videos to be private and associate with user's personal tenant
-- All existing videos are already private by default (DEFAULT TRUE above)
-- Now we need to associate them with the user's personal tenant

-- Update existing videos to use the user's personal tenant
-- We need to join with the userservice database to get personal tenant IDs
UPDATE videos 
SET tenant_id = (
    SELECT t.id 
    FROM userservice_tenants t 
    JOIN userservice_tenant_users tu ON t.id = tu.tenant_id 
    WHERE t.is_personal = TRUE 
    AND tu.user_id = videos.uploaded_user_id
    AND tu.role = 'super_admin'
    LIMIT 1
)
WHERE tenant_id IS NULL; 