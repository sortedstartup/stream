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
        return <XCircle className="h-4 w-4" />
      case 'danger':
        return <AlertCircle className="h-4 w-4" />
      case 'warning':
        return <AlertTriangle className="h-4 w-4" />
      default:
        return <AlertTriangle className="h-4 w-4" />
    }
  }

  const handleUpgradeClick = () => {
    // Navigate to billing page or trigger upgrade modal
    window.location.href = '/billing'
  }

  return (
    <div className={`border rounded px-3 py-2 ${getAlertStyles(level)} ${className}`}>
      <div className="flex items-center justify-between">
        <div className="flex items-center">
          <div className="flex-shrink-0 mr-2">
            {getIcon(level)}
          </div>
          <div className="text-sm">
            <span className="font-medium mr-2">
              {level === 'critical' ? 'Storage Full' : 'Storage Warning'}:
            </span>
            <span>
              {primaryWarning.type === 'storage' 
                ? `${formatStorageUsed(subscription.usage.storage_used_bytes || 0)} / ${formatStorageLimit(subscription.plan.storage_limit_bytes || 0)} (${primaryWarning.percent.toFixed(1)}%)`
                : `${subscription.usage.users_count || 0} / ${subscription.plan.users_limit || 0} users (${primaryWarning.percent.toFixed(1)}%)`
              }
            </span>
          </div>
        </div>
        
        {/* Compact upgrade button for free users */}
        {subscription.plan.id === 'free' && (
          <button
            onClick={handleUpgradeClick}
            className="btn btn-xs btn-primary ml-3"
          >
            Upgrade
          </button>
        )}
      </div>
    </div>
  )
} 