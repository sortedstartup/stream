-- Generic Payment Service Queries (Application Agnostic)

-- Plan queries
-- name: GetPlan :one
SELECT * FROM paymentservice_plans WHERE id = ? LIMIT 1;

-- name: GetActivePlans :many
SELECT * FROM paymentservice_plans WHERE is_active = TRUE ORDER BY price_cents ASC;

-- User subscription queries (core functionality)
-- name: GetUserSubscription :one
SELECT * FROM paymentservice_user_subscriptions WHERE user_id = ? LIMIT 1;

-- name: CreateUserSubscription :one
INSERT INTO paymentservice_user_subscriptions (
    id, user_id, plan_id, provider, provider_customer_id, 
    provider_subscription_id, status, current_period_start, 
    current_period_end, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateUserSubscriptionStatus :exec
UPDATE paymentservice_user_subscriptions 
SET status = ?, current_period_start = ?, current_period_end = ?, updated_at = ?
WHERE user_id = ?;

-- name: UpdateUserSubscriptionProvider :exec
UPDATE paymentservice_user_subscriptions 
SET provider_customer_id = ?, provider_subscription_id = ?, updated_at = ?
WHERE user_id = ?;

-- User usage queries (core functionality)
-- name: GetUserUsage :one
SELECT * FROM paymentservice_user_usage WHERE user_id = ? LIMIT 1;

-- name: CreateUserUsage :one
INSERT INTO paymentservice_user_usage (
    user_id, storage_used_bytes, users_count,
    last_calculated_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateUserUsage :exec
UPDATE paymentservice_user_usage 
SET storage_used_bytes = ?, users_count = ?,
    last_calculated_at = ?, updated_at = ?
WHERE user_id = ?;

-- name: UpdateUserStorageUsage :exec
UPDATE paymentservice_user_usage 
SET storage_used_bytes = storage_used_bytes + ?, last_calculated_at = ?, updated_at = ?
WHERE user_id = ?;

-- name: UpdateUserUsersCount :exec
UPDATE paymentservice_user_usage 
SET users_count = users_count + ?, last_calculated_at = ?, updated_at = ?
WHERE user_id = ?;

-- Core access check query (main functionality)
-- name: CheckUserAccess :one
SELECT 
    -- Subscription info
    s.user_id,
    s.plan_id,
    s.status as subscription_status,
    
    -- Plan limits
    p.storage_limit_bytes,
    p.users_limit,
    p.name as plan_name,
    p.price_cents,
    
    -- Current usage
    COALESCE(u.storage_used_bytes, 0) as current_storage_bytes,
    COALESCE(u.users_count, 0) as current_users_count,
    
    -- Access decision for storage
    CASE 
        WHEN s.status = 'active' 
         AND COALESCE(u.storage_used_bytes, 0) < p.storage_limit_bytes
        THEN 1 
        ELSE 0 
    END as has_storage_access,
    
    -- Access decision for users
    CASE 
        WHEN s.status = 'active' 
         AND COALESCE(u.users_count, 0) < p.users_limit
        THEN 1 
        ELSE 0 
    END as has_users_access,
    
    -- Usage percentages for warnings
    CAST(COALESCE(u.storage_used_bytes, 0) AS FLOAT) / p.storage_limit_bytes * 100 as storage_usage_percent,
    CAST(COALESCE(u.users_count, 0) AS FLOAT) / p.users_limit * 100 as users_usage_percent

FROM paymentservice_user_subscriptions s
LEFT JOIN paymentservice_plans p ON s.plan_id = p.id  
LEFT JOIN paymentservice_user_usage u ON s.user_id = u.user_id
WHERE s.user_id = ?; 