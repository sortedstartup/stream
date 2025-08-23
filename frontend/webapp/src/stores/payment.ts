import { atom } from 'nanostores'
import { $authToken, $currentUser } from '../auth/store/auth'
import { 
  PaymentServiceClient, 
  GetUserSubscriptionRequest,
  GetPlansRequest,
  GetPlansResponse,
  CreateCheckoutSessionRequest,
  UserSubscriptionInfo,
  Plan,
  UserUsage,
  Subscription
} from '../proto/paymentservice'
import { Timestamp } from '../proto/google/protobuf/timestamp'
import { useStore } from '@nanostores/react'

// Payment state atoms
export const $userSubscription = atom<UserSubscriptionInfo | null>(null)
export const $isLoadingSubscription = atom<boolean>(false)
export const $subscriptionError = atom<string | null>(null)
export const $isCreatingCheckout = atom<boolean>(false)

// Plans state atoms
export const $availablePlans = atom<Plan[]>([])
export const $isLoadingPlans = atom<boolean>(false)
export const $plansError = atom<string | null>(null)

const apiUrl = import.meta.env.VITE_PUBLIC_API_URL?.replace(/\/$/, "")

// Initialize the PaymentService client with proper configuration
const paymentServiceClient = new PaymentServiceClient(
  apiUrl,
  {},
  {
    unaryInterceptors: [
      {
        intercept: (request, invoker) => {
          const metadata = request.getMetadata();
          metadata["authorization"] = $authToken.get();
          return invoker(request);
        },
      },
    ],
  }
)

// Load user subscription information
export const loadUserSubscription = async () => {
  const currentUser = $currentUser.get()
  if (!currentUser?.uid) {
    // Don't set error if user is simply not loaded yet
    console.log('Payment: Current user not available yet, skipping subscription load')
    return
  }

  try {
    $isLoadingSubscription.set(true)
    $subscriptionError.set(null)
    
    const request = new GetUserSubscriptionRequest()
    request.user_id = currentUser.uid
    
    const response = await paymentServiceClient.GetUserSubscription(request, {})
    
    if (response.success && response.subscription_info) {
      $userSubscription.set(response.subscription_info)
    } else if (response.error_message === "No subscription found") {
      // This is normal for new users - they may not be initialized yet
      $subscriptionError.set("No subscription found - user may need to be initialized")
    } else {
      const errorMsg = response.error_message || 'Failed to load subscription'
      $subscriptionError.set(errorMsg)
      console.error('Payment: Subscription loading failed:', errorMsg)
    }
  } catch (error) {
    const errorMsg = 'Failed to load subscription information'
    $subscriptionError.set(errorMsg)
    console.error('Payment subscription loading error:', error)
  } finally {
    $isLoadingSubscription.set(false)
  }
}

// Create Stripe checkout session for plan upgrade
export const createCheckoutSession = async (planId: string) => {
  const currentUser = $currentUser.get()
  
  if (!currentUser?.uid) {
    console.error('Payment: No current user available for checkout')
    throw new Error('Please wait for user data to load, then try again')
  }

  try {
    $isCreatingCheckout.set(true)
    
    
    const request = new CreateCheckoutSessionRequest()
    request.user_id = currentUser.uid
    request.plan_id = planId
    request.success_url = `${window.location.origin}/billing/success`
    request.cancel_url = `${window.location.origin}/billing`
    
    const response = await paymentServiceClient.CreateCheckoutSession(request, {})
    
    if (response.success && response.checkout_url) {
      // Redirect to Stripe checkout
      window.location.href = response.checkout_url
      return response
    } else {
      const errorMsg = response.error_message || 'Failed to create checkout session'
      console.error('Payment: Checkout creation failed:', errorMsg)
      throw new Error(errorMsg)
    }
  } catch (error) {
    console.error('Checkout session creation error:', error)
    throw error
  } finally {
    $isCreatingCheckout.set(false)
  }
}

// Load available subscription plans
export const loadPlans = async () => {
  try {
    $isLoadingPlans.set(true)
    $plansError.set(null)
    
    const request = new GetPlansRequest()
    const response = await paymentServiceClient.GetPlans(request, {})
    
    if (response.success && response.plans) {
      $availablePlans.set(response.plans)
    } else {
      const errorMsg = response.error_message || 'Failed to load plans'
      $plansError.set(errorMsg)
      console.error('Payment: Plans loading failed:', errorMsg)
    }
  } catch (error) {
    const errorMsg = 'Failed to load subscription plans'
    $plansError.set(errorMsg)
    console.error('Payment plans loading error:', error)
  } finally {
    $isLoadingPlans.set(false)
  }
}

// Helper functions for subscription info
export const isFreePlan = (subscription: UserSubscriptionInfo | null): boolean => {
  return subscription?.plan?.id === 'free' || !subscription?.plan
}

export const isPaidPlan = (subscription: UserSubscriptionInfo | null): boolean => {
  return subscription?.plan?.id !== 'free' && subscription?.subscription?.status === 'active'
}

export const getStorageUsagePercent = (subscription: UserSubscriptionInfo | null): number => {
  return subscription?.usage?.storage_usage_percent || 0
}

export const getUsersUsagePercent = (subscription: UserSubscriptionInfo | null): number => {
  return subscription?.usage?.users_usage_percent || 0
}

export const formatStorageUsed = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export const formatStorageLimit = (bytes: number): string => {
  if (bytes >= 1024 * 1024 * 1024) {
    return Math.round(bytes / (1024 * 1024 * 1024)) + 'GB'
  }
  if (bytes >= 1024 * 1024) {
    return Math.round(bytes / (1024 * 1024)) + 'MB'
  }
  return Math.round(bytes / 1024) + 'KB'
}

// Usage warning helpers
export const getUsageWarningLevel = (usagePercent: number): 'none' | 'warning' | 'danger' | 'critical' => {
  if (usagePercent >= 100) return 'critical'
  if (usagePercent >= 90) return 'danger'
  if (usagePercent >= 75) return 'warning'
  return 'none'
}

export const getUsageWarningMessage = (usageType: 'storage' | 'users', usagePercent: number): string => {
  const level = getUsageWarningLevel(usagePercent)
  const typeLabel = usageType === 'storage' ? 'Storage' : 'User limit'
  
  switch (level) {
    case 'critical':
      return `${typeLabel} limit reached (100%). Please upgrade your plan to continue.`
    case 'danger':
      return `${typeLabel} ${usagePercent.toFixed(1)}% full. Consider upgrading your plan.`
    case 'warning':
      return `${typeLabel} ${usagePercent.toFixed(1)}% full. Consider upgrading your plan.`
    default:
      return ''
  }
} 