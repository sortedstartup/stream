import React from 'react'
import { useEffect, useRef } from 'react'
import { useNavigate } from 'react-router'
import { useStore } from '@nanostores/react'
import { $isLoggedIn } from '../store/auth'
import { startUi, signOut } from '../providers/firebase-auth'
import { setAuthState, clearAuthState } from '../store/auth'
import { Header } from '../../components/layout/Header'
import { createUserIfNotExists } from '../../stores/users'

export const LoginPage = () => {
    
  const navigate = useNavigate()
  const isLoggedIn = useStore($isLoggedIn)
  const firebaseUiContainerRef = useRef(null)

  useEffect(() => {
    if (!isLoggedIn && firebaseUiContainerRef.current) {
      startUi('#firebaseui-auth-container', handleAuthSuccess)
    }
  }, [isLoggedIn])

  const handleAuthSuccess = async (authResult) => {
    const { user } = authResult
    const token = await user.getIdToken()

    setAuthState({
      user,  // Pass the Firebase user object directly
      token
    })

    try {
      await createUserIfNotExists()
      // Reload tenants after user creation to ensure personal tenant is loaded
      const { loadUserTenants } = await import('../../stores/tenants')
      await loadUserTenants()
    } catch (error) {
      await signOut()
    }

    navigate('/')
  }

  const handleLogout = async () => {
    await signOut()
    clearAuthState()
    navigate('/login')
  }




  if (isLoggedIn) {
    return (
      <>
        <Header/>
        <div className="min-h-[calc(100vh-4rem)] hero bg-base-200">
          <div className="hero-content text-center">
            <div className="card w-96 bg-base-100 shadow-xl">
              <div className="card-body">
                <h2 className="card-title justify-center text-2xl">You are logged in</h2>
                <div className="card-actions justify-center mt-6">
                  <button
                    onClick={handleLogout}
                    className="btn btn-error"
                  >
                    Logout
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </>
    )
  }

  return (
    <>
      <Header/>
      <div className="min-h-[calc(100vh-4rem)] hero bg-base-200">
        <div className="hero-content">
          <div className="card w-96 bg-base-100 shadow-xl">
            <div className="card-body">
              <h2 className="card-title justify-center text-2xl">Login</h2>
              <div id="firebaseui-auth-container" ref={firebaseUiContainerRef}></div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
} 