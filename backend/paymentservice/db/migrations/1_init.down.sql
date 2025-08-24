-- Drop payment service tables in reverse order (Generic)
DROP INDEX IF EXISTS idx_user_subscriptions_status;
DROP INDEX IF EXISTS idx_user_subscriptions_user_id;

DROP TABLE IF EXISTS paymentservice_user_subscriptions;
DROP TABLE IF EXISTS paymentservice_plans; 