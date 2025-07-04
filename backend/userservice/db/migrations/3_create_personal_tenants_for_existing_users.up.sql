-- Migration to create personal tenants for existing users who don't have one
-- This is a one-time migration to ensure all existing users have personal tenants

-- Insert personal tenants for users who don't already have one
INSERT INTO userservice_tenants (
    id,
    name,
    description,
    is_personal,
    created_at,
    created_by
)
SELECT 
    lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))) as id,
    u.username as name,
    'Personal workspace' as description,
    true as is_personal,
    datetime('now') as created_at,
    u.id as created_by
FROM userservice_users u
WHERE NOT EXISTS (
    SELECT 1 
    FROM userservice_tenants t 
    JOIN userservice_tenant_users tu ON t.id = tu.tenant_id 
    WHERE tu.user_id = u.id AND t.is_personal = true
);

-- Add users to their personal tenants as super_admin
INSERT INTO userservice_tenant_users (
    id,
    tenant_id,
    user_id,
    role,
    created_at
)
SELECT 
    lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))) as id,
    t.id as tenant_id,
    u.id as user_id,
    'super_admin' as role,
    datetime('now') as created_at
FROM userservice_users u
JOIN userservice_tenants t ON t.created_by = u.id AND t.is_personal = true
WHERE NOT EXISTS (
    SELECT 1 
    FROM userservice_tenant_users tu 
    WHERE tu.tenant_id = t.id AND tu.user_id = u.id
); 