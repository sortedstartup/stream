CREATE TABLE userservice_tenants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    is_personal BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE for auto-created personal tenants
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL REFERENCES userservice_users(id)
);

CREATE TABLE userservice_tenant_users (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES userservice_tenants(id),
    user_id TEXT NOT NULL REFERENCES userservice_users(id),
    role TEXT NOT NULL DEFAULT 'member', -- Simple string: 'super_admin', 'admin', 'member'
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, user_id)
);