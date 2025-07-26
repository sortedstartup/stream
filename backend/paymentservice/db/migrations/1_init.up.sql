-- Payment service database schema - Model 2 (Pay Per User)
-- Supports multiple payment providers (Stripe, Razorpay)

-- Plan definitions table (generic for any application)
CREATE TABLE paymentservice_plans (
    id TEXT PRIMARY KEY,                    -- 'free', 'standard', 'premium'
    name TEXT NOT NULL,
    storage_limit_bytes INTEGER NOT NULL,  -- Storage limit in bytes
    api_calls_limit INTEGER NOT NULL,      -- API calls limit per period
    compute_hours_limit INTEGER NOT NULL,  -- Compute hours limit per period
    price_cents INTEGER DEFAULT 0,         -- 0 for free, provider price for paid
    provider_price_id TEXT,                -- Provider-specific price ID (Stripe/Razorpay)
    is_active BOOLEAN DEFAULT TRUE,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- User subscriptions table (generic: one subscription per user)
CREATE TABLE paymentservice_user_subscriptions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,          -- User who paid
    plan_id TEXT NOT NULL DEFAULT 'free',
    provider TEXT NOT NULL DEFAULT 'stripe', -- 'stripe' or 'razorpay'
    provider_customer_id TEXT,             -- Provider customer ID
    provider_subscription_id TEXT,         -- Provider subscription ID
    status TEXT NOT NULL DEFAULT 'active', -- active, canceled, past_due, incomplete
    current_period_start INTEGER,
    current_period_end INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    
    FOREIGN KEY (plan_id) REFERENCES paymentservice_plans(id)
);

-- User usage tracking table (generic usage metrics)
CREATE TABLE paymentservice_user_usage (
    user_id TEXT PRIMARY KEY,
    storage_used_bytes INTEGER DEFAULT 0,   -- Total storage used by user
    api_calls_used INTEGER DEFAULT 0,       -- Total API calls used in current period
    compute_hours_used INTEGER DEFAULT 0,   -- Total compute hours used in current period
    last_calculated_at INTEGER,            -- Last time usage was calculated
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    
    FOREIGN KEY (user_id) REFERENCES paymentservice_user_subscriptions(user_id)
);

-- Indexes for better performance
CREATE INDEX idx_user_subscriptions_user_id ON paymentservice_user_subscriptions(user_id);
CREATE INDEX idx_user_subscriptions_status ON paymentservice_user_subscriptions(status);
CREATE INDEX idx_user_usage_user_id ON paymentservice_user_usage(user_id);

-- Insert default plans (for Stream application - can be customized per app)
INSERT INTO paymentservice_plans (id, name, storage_limit_bytes, api_calls_limit, compute_hours_limit, price_cents, created_at, updated_at) VALUES
('free', 'Free Plan', 1073741824, 1000, 10, 0, strftime('%s', 'now'), strftime('%s', 'now')),
('standard', 'Standard Plan', 107374182400, 50000, 500, 2900, strftime('%s', 'now'), strftime('%s', 'now')); 