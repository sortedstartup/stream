import React, { useEffect, useState } from 'react'
import { useStore } from '@nanostores/react'
import { CreditCard, Users, HardDrive, CheckCircle, Clock } from 'react-feather'
import { 
  $userSubscription, 
  $isLoadingSubscription, 
  $subscriptionError,
  $isCreatingCheckout,
  loadUserSubscription,
  createCheckoutSession,
  isFreePlan,
  isPaidPlan,
  formatStorageUsed,
  formatStorageLimit,
  getStorageUsagePercent,
  getUsersUsagePercent
} from '../stores/payment'
import { showSuccessToast, showErrorToast } from '../utils/toast'
import { UsageWarningBanner } from '../components/UsageWarningBanner'
import { $currentUser, $authInitialized } from '../auth/store/auth'
import { Layout } from '../components/layout/Layout'

export const BillingPage = () => {
  const subscription = useStore($userSubscription)
  const isLoading = useStore($isLoadingSubscription)
  const error = useStore($subscriptionError)
  const isCreatingCheckout = useStore($isCreatingCheckout)
  const currentUser = useStore($currentUser)
  const authInitialized = useStore($authInitialized)
  
  const [upgradeError, setUpgradeError] = useState('')
  
  // Check if user data is ready for checkout - wait for auth to initialize
  const isUserReady = authInitialized && Boolean(currentUser?.uid)

  useEffect(() => {
    // Load subscription when billing page is accessed and auth is ready
    if (authInitialized && currentUser?.uid) {
      loadUserSubscription()
    }
  }, [authInitialized, currentUser?.uid])

  // Show loading if auth is not yet initialized
  if (!authInitialized) {
    return (
      <Layout>
        <div className="min-h-screen bg-gray-50 py-8">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex items-center justify-center">
              <span className="loading loading-spinner loading-lg"></span>
              <span className="ml-2">Initializing authentication...</span>
            </div>
          </div>
        </div>
      </Layout>
    )
  }

  const handleUpgrade = async (planId) => {
    try {
      setUpgradeError('')
      await createCheckoutSession(planId)
      // User will be redirected to Stripe checkout
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to start upgrade process'
      setUpgradeError(errorMessage)
      showErrorToast(errorMessage)
    }
  }

  if (isLoading) {
    return (
      <Layout>
        <div className="min-h-screen bg-gray-50 py-8">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex items-center justify-center">
              <span className="loading loading-spinner loading-lg"></span>
              <span className="ml-2">Loading subscription details...</span>
            </div>
          </div>
        </div>
      </Layout>
    )
  }

  if (error) {
    return (
      <Layout>
        <div className="min-h-screen bg-gray-50 py-8">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="bg-red-50 border border-red-200 rounded-lg p-4">
              <div className="text-red-800">
                <h3 className="font-medium">Error Loading Subscription</h3>
                <p className="mt-1 text-sm">{error}</p>
                <button 
                  onClick={loadUserSubscription}
                  className="mt-3 btn btn-sm btn-outline btn-error"
                >
                  Retry
                </button>
              </div>
            </div>
          </div>
        </div>
      </Layout>
    )
  }

  const currentPlan = subscription?.plan
  const usage = subscription?.usage
  const storagePercent = getStorageUsagePercent(subscription)
  const usersPercent = getUsersUsagePercent(subscription)

  return (
    <Layout>
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-gray-900">Billing & Usage</h1>
            <p className="mt-2 text-gray-600">Manage your subscription and monitor usage</p>
          </div>

          {/* Usage Warning Banner */}
          <UsageWarningBanner subscription={subscription} className="mb-6" />

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Current Plan */}
            <div className="lg:col-span-2">
              <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                <div className="flex items-center justify-between mb-6">
                  <h2 className="text-xl font-semibold text-gray-900">Current Plan</h2>
                  {isPaidPlan(subscription) && (
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                      <CheckCircle className="w-3 h-3 mr-1" />
                      Active
                    </span>
                  )}
                </div>

                {currentPlan && (
                  <div className="border rounded-lg p-4 mb-6">
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="text-lg font-medium text-gray-900">{currentPlan.name}</h3>
                        <p className="text-gray-600">
                          ${(currentPlan.price_cents / 100).toFixed(2)}/month
                        </p>
                      </div>
                      <div className="text-right">
                        <div className="text-sm text-gray-600">
                          <div className="flex items-center">
                            <HardDrive className="w-4 h-4 mr-1" />
                            {formatStorageLimit(currentPlan.storage_limit_bytes)}
                          </div>
                          <div className="flex items-center mt-1">
                            <Users className="w-4 h-4 mr-1" />
                            {currentPlan.users_limit} users
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {/* Usage Statistics */}
                <div className="space-y-4">
                  <h3 className="text-lg font-medium text-gray-900">Usage</h3>
                  
                  {/* Storage Usage */}
                  <div className="border rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center">
                        <HardDrive className="w-5 h-5 text-gray-500 mr-2" />
                        <span className="font-medium">Storage</span>
                      </div>
                      <span className="text-sm text-gray-600">
                        {formatStorageUsed(usage?.storage_used_bytes || 0)} / {formatStorageLimit(currentPlan?.storage_limit_bytes || 0)}
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div 
                        className={`h-2 rounded-full ${storagePercent >= 90 ? 'bg-red-500' : storagePercent >= 75 ? 'bg-orange-500' : 'bg-blue-500'}`}
                        style={{ width: `${Math.min(storagePercent, 100)}%` }}
                      ></div>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">{storagePercent.toFixed(1)}% used</p>
                  </div>

                  {/* Users Usage */}
                  <div className="border rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center">
                        <Users className="w-5 h-5 text-gray-500 mr-2" />
                        <span className="font-medium">Users</span>
                      </div>
                      <span className="text-sm text-gray-600">
                        {usage?.users_count || 0} / {currentPlan?.users_limit || 0}
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div 
                        className={`h-2 rounded-full ${usersPercent >= 90 ? 'bg-red-500' : usersPercent >= 75 ? 'bg-orange-500' : 'bg-blue-500'}`}
                        style={{ width: `${Math.min(usersPercent, 100)}%` }}
                      ></div>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">{usersPercent.toFixed(1)}% used</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Upgrade Plan Card */}
            {isFreePlan(subscription) && (
              <div className="lg:col-span-1">
                <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">Upgrade to Standard</h3>
                  
                  <div className="space-y-4 mb-6">
                    <div className="flex items-center">
                      <CheckCircle className="w-5 h-5 text-green-500 mr-2" />
                      <span className="text-sm">100GB Storage</span>
                    </div>
                    <div className="flex items-center">
                      <CheckCircle className="w-5 h-5 text-green-500 mr-2" />
                      <span className="text-sm">50 Users</span>
                    </div>
                    <div className="flex items-center">
                      <CheckCircle className="w-5 h-5 text-green-500 mr-2" />
                      <span className="text-sm">Unlimited Workspaces</span>
                    </div>
                    <div className="flex items-center">
                      <CheckCircle className="w-5 h-5 text-green-500 mr-2" />
                      <span className="text-sm">Priority Support</span>
                    </div>
                  </div>

                  <div className="text-center mb-4">
                    <div className="text-3xl font-bold text-gray-900">$29</div>
                    <div className="text-sm text-gray-600">per month</div>
                  </div>

                  {upgradeError && (
                    <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-700 text-sm">
                      {upgradeError}
                    </div>
                  )}      

                  <button
                    onClick={() => handleUpgrade('standard')}
                    disabled={isCreatingCheckout || !isUserReady}
                    className="w-full btn btn-primary"
                  >
                    {isCreatingCheckout ? (
                      <>
                        <span className="loading loading-spinner loading-sm"></span>
                        Starting checkout...
                      </>
                    ) : !authInitialized ? (
                      <>
                        <span className="loading loading-spinner loading-sm"></span>
                        Initializing...
                      </>
                    ) : !currentUser?.uid ? (
                      <>
                        <span className="loading loading-spinner loading-sm"></span>
                        Please log in
                      </>
                    ) : (
                      <>
                        <CreditCard className="w-4 h-4 mr-2" />
                        Upgrade Now
                      </>
                    )}
                  </button>

                  <p className="text-xs text-gray-500 text-center mt-3">
                    Secure payment powered by Stripe
                  </p>
                </div>
              </div>
            )}

            {/* Paid Plan Info */}
            {isPaidPlan(subscription) && (
              <div className="lg:col-span-1">
                <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">Subscription Status</h3>
                  
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Status:</span>
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                        <CheckCircle className="w-3 h-3 mr-1" />
                        {subscription?.subscription?.status || 'Active'}
                      </span>
                    </div>
                    
                    {subscription?.subscription?.current_period_end && (
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">Next billing:</span>
                        <span className="text-sm font-medium">
                          {new Date(subscription.subscription.current_period_end.seconds * 1000).toLocaleDateString()}
                        </span>
                      </div>
                    )}
                  </div>

                  <div className="mt-6 p-4 bg-blue-50 rounded-lg">
                    <div className="flex items-center">
                      <Clock className="w-5 h-5 text-blue-500 mr-2" />
                      <span className="text-sm text-blue-700">
                        Manage billing and invoices through your Stripe dashboard
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </Layout>
  )
} 