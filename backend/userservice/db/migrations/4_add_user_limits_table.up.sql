-- Add user limits tracking table to user service
-- This replaces the user count tracking that was previously in payment service

CREATE TABLE userservice_user_limits (
    user_id TEXT PRIMARY KEY,
    users_count INTEGER DEFAULT 0,          -- Total users managed by this user across all tenants
    last_calculated_at INTEGER,            -- Last time usage was calculated
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Index for better performance
CREATE INDEX idx_user_limits_user_id ON userservice_user_limits(user_id); 