-- Generic Payment Service Queries (Application Agnostic)

-- Plan queries
-- name: GetPlan :one
SELECT * FROM paymentservice_plans WHERE id = ? LIMIT 1;

-- name: GetActivePlans :many
SELECT * FROM paymentservice_plans WHERE is_active = TRUE ORDER BY price_cents ASC;

-- name: CreatePlan :one
INSERT INTO paymentservice_plans (
    id, name, price_cents, is_active, created_at, updated_at
) VALUES (?, ?, ?, TRUE, ?, ?)
RETURNING *;

-- name: UpdatePlan :one
UPDATE paymentservice_plans 
SET name = ?, price_cents = ?, updated_at = ?
WHERE id = ?
RETURNING *;

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

-- name: UpdateUserSubscriptionPlan :exec
UPDATE paymentservice_user_subscriptions 
SET plan_id = ?, updated_at = ?
WHERE user_id = ?;

-- name: GetUserByProviderSubscriptionID :one
SELECT user_id FROM paymentservice_user_subscriptions 
WHERE provider_subscription_id = ? LIMIT 1;

 