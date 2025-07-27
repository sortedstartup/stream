import React, { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { CheckCircle, ArrowRight } from 'react-feather'
import { loadUserSubscription } from '../stores/payment'
import { showSuccessToast } from '../utils/toast'
import { Layout } from '../components/layout/Layout'

export const BillingSuccessPage = () => {
  const navigate = useNavigate()

  useEffect(() => {
    // Reload subscription data after successful payment
    loadUserSubscription()
    
    // Show success message
    showSuccessToast('Payment successful! Your subscription has been upgraded.')
  }, [])

  const handleContinue = () => {
    navigate('/workspace')
  }

  const handleViewBilling = () => {
    navigate('/billing')
  }

  return (
    <Layout>
      <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full space-y-8">
          <div className="text-center">
            {/* Success Icon */}
            <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-green-100 mb-6">
              <CheckCircle className="h-10 w-10 text-green-600" />
            </div>
            
            {/* Success Message */}
            <h2 className="text-3xl font-bold text-gray-900 mb-2">
              Payment Successful!
            </h2>
            <p className="text-gray-600 mb-8">
              Your subscription has been upgraded successfully. You now have access to all premium features.
            </p>

            {/* Feature Highlights */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-8">
              <h3 className="text-lg font-medium text-gray-900 mb-4">You now have access to:</h3>
              <div className="space-y-3 text-left">
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 text-green-500 mr-3" />
                  <span className="text-sm text-gray-700">100GB storage space</span>
                </div>
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 text-green-500 mr-3" />
                  <span className="text-sm text-gray-700">Up to 50 team members</span>
                </div>
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 text-green-500 mr-3" />
                  <span className="text-sm text-gray-700">Unlimited workspaces</span>
                </div>
                <div className="flex items-center">
                  <CheckCircle className="h-4 w-4 text-green-500 mr-3" />
                  <span className="text-sm text-gray-700">Priority customer support</span>
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="space-y-3">
              <button
                onClick={handleContinue}
                className="w-full btn btn-primary"
              >
                Continue to Workspace
                <ArrowRight className="ml-2 h-4 w-4" />
              </button>
              
              <button
                onClick={handleViewBilling}
                className="w-full btn btn-outline"
              >
                View Billing Details
              </button>
            </div>

            {/* Receipt Note */}
            <p className="text-xs text-gray-500 mt-6">
              A receipt has been sent to your email address. You can manage your subscription anytime from the billing page.
            </p>
          </div>
        </div>
      </div>
    </Layout>
  )
} 