import React from 'react'
import { AlertTriangle, AlertCircle, XCircle } from 'react-feather'
import { UserSubscriptionInfo } from '../proto/paymentservice'
import { 
  getStorageUsagePercent, 
  getUsersUsagePercent, 
  getUsageWarningLevel, 
  getUsageWarningMessage,
  formatStorageUsed,
  formatStorageLimit
} from '../stores/payment'

interface UsageWarningBannerProps {
  subscription: UserSubscriptionInfo | null
  className?: string
}

export const UsageWarningBanner: React.FC<UsageWarningBannerProps> = ({ 
  subscription, 
  className = "" 
}) => {
  if (!subscription?.usage || !subscription?.plan) {
    return null
  }

  const storagePercent = getStorageUsagePercent(subscription)
  const usersPercent = getUsersUsagePercent(subscription)
  const storageLevel = getUsageWarningLevel(storagePercent)
  const usersLevel = getUsageWarningLevel(usersPercent)

  // Only show if there are warnings
  if (storageLevel === 'none' && usersLevel === 'none') {
    return null
  }

  // Determine the highest severity level to show
  const showStorageWarning = storageLevel !== 'none'
  const showUsersWarning = usersLevel !== 'none'
  
  // Pick the most severe warning to display
  const criticalWarnings = []
  if (storageLevel === 'critical') criticalWarnings.push({ type: 'storage', percent: storagePercent })
  if (usersLevel === 'critical') criticalWarnings.push({ type: 'users', percent: usersPercent })
  
  const dangerWarnings = []
  if (storageLevel === 'danger') dangerWarnings.push({ type: 'storage', percent: storagePercent })
  if (usersLevel === 'danger') dangerWarnings.push({ type: 'users', percent: usersPercent })
  
  const warningWarnings = []
  if (storageLevel === 'warning') warningWarnings.push({ type: 'storage', percent: storagePercent })
  if (usersLevel === 'warning') warningWarnings.push({ type: 'users', percent: usersPercent })

  let primaryWarning
  let level: 'warning' | 'danger' | 'critical' = 'warning'
  
  if (criticalWarnings.length > 0) {
    primaryWarning = criticalWarnings[0]
    level = 'critical'
  } else if (dangerWarnings.length > 0) {
    primaryWarning = dangerWarnings[0]
    level = 'danger'
  } else if (warningWarnings.length > 0) {
    primaryWarning = warningWarnings[0]
    level = 'warning'
  }

  if (!primaryWarning) {
    return null
  }

  const getAlertStyles = (level: string) => {
    switch (level) {
      case 'critical':
        return 'border-red-200 bg-red-50 text-red-800'
      case 'danger':
        return 'border-orange-200 bg-orange-50 text-orange-800'
      case 'warning':
        return 'border-yellow-200 bg-yellow-50 text-yellow-800'
      default:
        return 'border-blue-200 bg-blue-50 text-blue-800'
    }
  }

  const getIcon = (level: string) => {
    switch (level) {
      case 'critical':
        return <XCircle className="h-5 w-5" />
      case 'danger':
        return <AlertCircle className="h-5 w-5" />
      case 'warning':
        return <AlertTriangle className="h-5 w-5" />
      default:
        return <AlertTriangle className="h-5 w-5" />
    }
  }

  const handleUpgradeClick = () => {
    // Navigate to billing page or trigger upgrade modal
    window.location.href = '/billing'
  }

  return (
    <div className={`border rounded-lg p-4 ${getAlertStyles(level)} ${className}`}>
      <div className="flex items-start">
        <div className="flex-shrink-0">
          {getIcon(level)}
        </div>
        <div className="ml-3 flex-1">
          <h3 className="text-sm font-medium">
            {level === 'critical' ? 'Action Required' : 'Usage Warning'}
          </h3>
          <div className="mt-2 text-sm">
            <p>{getUsageWarningMessage(primaryWarning.type as 'storage' | 'users', primaryWarning.percent)}</p>
            
            {/* Usage details */}
            <div className="mt-3 space-y-2">
              {showStorageWarning && (
                <div className="flex items-center justify-between text-xs">
                  <span>Storage:</span>
                  <span>
                    {formatStorageUsed(subscription.usage.storage_used_bytes || 0)} / {formatStorageLimit(subscription.plan.storage_limit_bytes || 0)} 
                    ({storagePercent.toFixed(1)}%)
                  </span>
                </div>
              )}
              {showUsersWarning && (
                <div className="flex items-center justify-between text-xs">
                  <span>Users:</span>
                  <span>
                    {subscription.usage.users_count || 0} / {subscription.plan.users_limit || 0} 
                    ({usersPercent.toFixed(1)}%)
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>
        
        {/* Upgrade button for free users */}
        {subscription.plan.id === 'free' && (
          <div className="ml-4 flex-shrink-0">
            <button
              onClick={handleUpgradeClick}
              className="btn btn-sm btn-primary"
            >
              Upgrade Plan
            </button>
          </div>
        )}
      </div>
    </div>
  )
} 