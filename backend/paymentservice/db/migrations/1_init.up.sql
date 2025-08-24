-- Payment service database schema (Pure Payment Functionality)
-- Supports multiple payment providers (Stripe, Razorpay)

-- Plan definitions table (generic for any application)
CREATE TABLE paymentservice_plans (
    id TEXT PRIMARY KEY,                    -- 'free', 'standard', 'premium'
    name TEXT NOT NULL,
    price_cents INTEGER DEFAULT 0,         -- 0 for free, provider price for paid
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

-- Indexes for better performance
CREATE INDEX idx_user_subscriptions_user_id ON paymentservice_user_subscriptions(user_id);
CREATE INDEX idx_user_subscriptions_status ON paymentservice_user_subscriptions(status); 