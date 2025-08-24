-- Add storage usage tracking table to video service
-- This replaces the usage tracking that was previously in payment service

CREATE TABLE videoservice_user_storage_usage (
    user_id TEXT PRIMARY KEY,
    storage_used_bytes INTEGER DEFAULT 0,   -- Total storage used by user
    last_calculated_at INTEGER,            -- Last time usage was calculated
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Index for better performance
CREATE INDEX idx_storage_usage_user_id ON videoservice_user_storage_usage(user_id); 