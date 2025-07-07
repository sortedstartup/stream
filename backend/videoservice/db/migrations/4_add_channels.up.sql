-- Channels are organizational tools within tenants
-- Note: This creates channel tables in videoservice database
-- User/tenant references are by ID only (no FK constraints due to cross-database)

-- Channels table
CREATE TABLE channels (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL, -- References userservice_tenants(id) but no FK constraint
    name TEXT NOT NULL,
    description TEXT,
    created_by TEXT NOT NULL, -- References userservice_users(id) but no FK constraint
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Channel members table  
CREATE TABLE channel_members (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL, -- References userservice_users(id) but no FK constraint
    role TEXT NOT NULL DEFAULT 'viewer', -- Simple string: 'owner', 'uploader', 'viewer'
    added_by TEXT NOT NULL, -- References userservice_users(id) but no FK constraint
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(channel_id, user_id)
);

-- Indexes for performance and security (always filter by tenant_id)
CREATE INDEX idx_channels_tenant ON channels(tenant_id);
CREATE INDEX idx_channels_created_by ON channels(created_by);
CREATE INDEX idx_channels_tenant_created ON channels(tenant_id, created_at);
CREATE INDEX idx_channel_members_channel ON channel_members(channel_id);
CREATE INDEX idx_channel_members_user ON channel_members(user_id);
CREATE INDEX idx_channel_members_role ON channel_members(role); 