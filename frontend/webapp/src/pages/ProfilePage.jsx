import React from 'react'
import { useStore } from '@nanostores/react'
import { $currentUser } from '../auth/store/auth'
import { Header } from '../components/layout/Header'

export const ProfilePage = () => {
  const currentUser = useStore($currentUser)

  return (
    <div className="min-h-screen bg-base-200">
      <Header />
      <div className="container mx-auto px-4 py-8">
        <div className="bg-base-100 shadow-xl rounded-lg p-6 max-w-2xl mx-auto">
          <h1 className="text-2xl font-bold mb-6">Profile</h1>
          
          <div className="space-y-4">
            <div className="flex items-center space-x-4">
              <div className="avatar">
                <div className="w-24 rounded-full">
                  {currentUser?.photoURL ? (
                    <img src={currentUser.photoURL} alt="Profile" />
                  ) : (
                    <div className="bg-primary text-primary-content rounded-full w-24 h-24 flex items-center justify-center text-2xl">
                      {currentUser?.displayName?.[0] || 'U'}
                    </div>
                  )}
                </div>
              </div>
              
              <div>
                <h2 className="text-xl font-semibold">
                  {currentUser?.displayName || 'User'}
                </h2>
                <p className="text-base-content/70">
                  {currentUser?.email}
                </p>
              </div>
            </div>

            <div className="divider"></div>

            <div className="space-y-2">
              <h3 className="text-lg font-semibold">Account Information</h3>
              <p>Email verified: {currentUser?.emailVerified ? 'Yes' : 'No'}</p>
              <p>Account created: {currentUser?.metadata?.creationTime}</p>
              <p>Last sign in: {currentUser?.metadata?.lastSignInTime}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
} 