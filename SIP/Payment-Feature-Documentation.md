# Payment Feature Documentation

## Overview
The Stream application now includes a freemium payment system that restricts user actions based on subscription plans. Users get free access with limits and can upgrade for expanded capabilities.

## Business Model

### Free Tier (Personal Workspace Only)
- **Storage**: 1GB maximum
- **Users**: 5 users maximum across all workspaces
- **Workspaces**: Personal workspace only (cannot create organizational workspaces)
- **Features**: Upload videos, add team members (within limits)

### Paid Tier - Standard Plan ($29/month)
- **Storage**: 100GB maximum
- **Users**: 50 users maximum across all workspaces
- **Workspaces**: Unlimited organizational workspaces + personal workspace
- **Features**: All free features + unlimited workspace creation

## How It Works

### User Journey

#### New Users
1. **Sign Up** → Automatically get free tier (1GB, 5 users)
2. **Personal Workspace** → Created automatically with free access
3. **Upload Videos** → Allowed until 1GB storage limit
4. **Add Users** → Allowed up to 5 users total
5. **Create Workspace** → Blocked (requires upgrade)

#### Existing Users (Pre-Payment)
1. **First Login** → Automatically initialized with free tier
2. **Same Limits Apply** → 1GB storage, 5 users, personal workspace only

#### Paid Users
1. **Upgrade** → Payment via Stripe checkout
2. **Expanded Limits** → 100GB storage, 50 users
3. **Create Workspaces** → Can create unlimited organizational workspaces
4. **Global Quota** → Limits apply across ALL owned workspaces

### Access Control Points

#### 1. Video Upload
- **Check**: Storage limit before upload
- **Block**: If upload would exceed user's storage quota
- **Track**: File size added to user's storage usage after successful upload
- **Error**: "Upload failed: Storage limit exceeded. Please upgrade your plan."

#### 2. Add Users to Workspaces
- **Check**: User limit before adding member
- **Logic**: Global limit across all workspaces owned by workspace creator
- **Block**: If adding user would exceed owner's user quota
- **Error**: "Cannot add user: User limit exceeded. Please upgrade your plan."

#### 3. Create New Workspace
- **Check**: User's subscription plan
- **Block**: Free users from creating organizational workspaces
- **Allow**: Only paid users can create organizational workspaces
- **Error**: "Workspace creation requires paid subscription. Please upgrade your plan."

### Warning System
- **75% Usage**: Yellow warning (logged)
- **90% Usage**: Orange warning (logged)
- **100% Usage**: Red warning + action blocked

## Technical Implementation

### Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Frontend      │    │   UserService    │    │ PaymentService  │
│                 │    │                  │    │                 │
│ • Upload UI     │───▶│ • User creation  │───▶│ • Subscription  │
│ • Workspace UI  │    │ • Workspace mgmt │    │ • Usage tracking│
│ • Warning UI    │    │ • Access checks  │    │ • Stripe integration│
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                         │
                                ▼                         │
                       ┌──────────────────┐              │
                       │   VideoService   │              │
                       │                  │              │
                       │ • Upload handler │──────────────┘
                       │ • Storage tracking│
                       └──────────────────┘
```

### Database Schema

The payment service tracks three main data types:
- **User Subscriptions**: Links users to their payment plans and Stripe subscription details
- **Usage Tracking**: Monitors storage usage and user count per subscription owner
- **Plan Definitions**: Defines available subscription tiers with their limits and pricing

### API Integration

#### gRPC Methods

1. **CheckUserAccess** - Validates if user can perform an action (upload, add user)
2. **UpdateUserUsage** - Records usage after successful actions (file upload, user addition)
3. **InitializeUser** - Sets up free subscription for new users
4. **GetUserSubscription** - Retrieves current subscription and usage details
5. **CreateCheckoutSession** - Generates Stripe payment URL for upgrades

### Service Integration

#### UserService Integration
- **User Signup**: Calls `InitializeUser` to setup free subscription
- **User Login**: Checks and initializes payment for existing users (migration)
- **Add User**: Calls `CheckUserAccess` before adding users to workspaces
- **Create Workspace**: Blocks free users from creating organizational workspaces

#### VideoService Integration
- **Upload Handler**: 
  1. Calls `CheckUserAccess` before allowing upload
  2. Blocks upload if storage limit would be exceeded
  3. Calls `UpdateUserUsage` after successful upload to track storage

### Configuration

The system requires:
- **Stripe API Keys**: Secret key, publishable key, webhook secret for payment processing
- **Plan Configuration**: Free tier (1GB, 5 users, $0) and Standard tier (100GB, 50 users, $29/month)
- **Price IDs**: Stripe-generated identifiers linking plans to payment products

## User Scenarios

### Scenario 1: Free User Uploads
1. Alice (free user) uploads 500MB video  **Allowed**
2. Alice uploads another 600MB video  **Blocked** (would exceed 1GB)
3. Error message: "Upload failed: Storage limit exceeded. Please upgrade your plan."

### Scenario 2: Free User Adds Team Members
1. Bob (free user) has personal workspace with 3 members  **Current**
2. Bob tries to add 2 more members  **Allowed** (would be 5 total)
3. Bob tries to add 1 more member  **Blocked** (would exceed 5 users)
4. Error message: "Cannot add user: User limit exceeded. Please upgrade your plan."

### Scenario 3: Free User Creates Workspace
1. Carol (free user) tries to create "Marketing Team" workspace
2. **Blocked** before creation
3. Error message: "Workspace creation requires paid subscription. Please upgrade your plan."

### Scenario 4: Paid User Usage
1. Dave (paid user) has 2 workspaces with 25 users each **Allowed** (50 total)
2. Dave uploads 50GB across all workspaces  **Allowed** (under 100GB)
3. Dave creates new "Sales Team" workspace **Allowed** (unlimited workspaces)
4. Dave tries to add 1 more user **Blocked** (would exceed 50 user limit)

### Scenario 5: Payment Upgrade
1. Eve (free user) hits storage limit
2. Eve clicks "Upgrade" button → Redirected to Stripe checkout
3. Eve completes payment → Webhook updates subscription to "standard"
4. Eve can now upload more videos (100GB limit) and create workspaces

## Error Messages

### Storage Limits
- `"Upload failed: Storage limit exceeded. Please upgrade your plan to continue uploading."`
- `"Upload failed: Your subscription is inactive. Please reactivate to continue uploading."`

### User Limits  
- `"Cannot add user: User limit exceeded. Please upgrade your plan to add more members."`
- `"Cannot add user: Subscription is inactive. Please reactivate to add members."`

### Workspace Creation
- `"Workspace creation requires paid subscription. Please upgrade your plan to create additional workspaces."`


## Warning Messages

### Storage Warnings
- **75%**: `"Storage 75.0% full. Consider upgrading your plan."`
- **90%**: `"Storage 90.0% full. Consider upgrading your plan."`

### User Warnings
- **75%**: `"Users 75.0% of limit. Consider upgrading your plan."`
- **90%**: `"Users 90.0% of limit. Consider upgrading your plan."`

## Migration Strategy

### Existing Users
- **Automatic Migration**: Existing users get free tier initialized on first login
- **No Manual Steps**: No database migration required
- **Self-Healing**: Works for any users we might miss
- **Gradual**: Users migrated as they log in

### Data Consistency
- **User Creation**: New users automatically get payment service initialized
- **Workspace Validation**: Organizational workspaces require paid subscriptions
- **Usage Tracking**: Video uploads automatically update storage usage



---

*This feature enables Stream to operate as a sustainable freemium SaaS product with clear upgrade incentives and usage-based limits.* 