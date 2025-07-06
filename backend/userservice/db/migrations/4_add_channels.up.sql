-- Add channel support to userservice
-- Channels are organizational tools within tenants

-- Channels table
CREATE TABLE userservice_channels (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES userservice_tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    is_private BOOLEAN NOT NULL DEFAULT TRUE,
    created_by TEXT NOT NULL REFERENCES userservice_users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Channel members table
CREATE TABLE userservice_channel_members (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL REFERENCES userservice_channels(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES userservice_users(id),
    role TEXT NOT NULL DEFAULT 'viewer' CHECK (role IN ('owner', 'uploader', 'viewer')),
    added_by TEXT NOT NULL REFERENCES userservice_users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(channel_id, user_id)
);

-- Indexes for performance and security
CREATE INDEX idx_channels_tenant ON userservice_channels(tenant_id);
CREATE INDEX idx_channels_created_by ON userservice_channels(created_by);
CREATE INDEX idx_channel_members_channel ON userservice_channel_members(channel_id);
CREATE INDEX idx_channel_members_user ON userservice_channel_members(user_id);
CREATE INDEX idx_channel_members_role ON userservice_channel_members(role); 