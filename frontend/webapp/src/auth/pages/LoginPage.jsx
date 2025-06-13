import React from 'react'
import { useEffect, useRef } from 'react'
import { useNavigate } from 'react-router'
import { useStore } from '@nanostores/react'
import { $isLoggedIn } from '../store/auth'
import { startUi, signOut } from '../providers/firebase-auth'
import { setAuthState, clearAuthState } from '../store/auth'
import { Header } from '../../components/layout/Header'

export const LoginPage = () => {
  const navigate = useNavigate()
  const isLoggedIn = useStore($isLoggedIn)
  const firebaseUiContainerRef = useRef(null)

  useEffect(() => {
    // If user is logged in, redirect to home
    if (isLoggedIn) {
      navigate('/')
      return
    }

    // Initialize Firebase UI if not logged in
    if (firebaseUiContainerRef.current) {
      startUi('#firebaseui-auth-container', handleAuthSuccess)
    }
  }, [isLoggedIn, navigate])

  const handleAuthSuccess = async (authResult) => {
    const { user } = authResult
    const token = await user.getIdToken()

    setAuthState({
      user: {
        id: user.uid,
        email: user.email,
        name: user.displayName,
        photoURL: user.photoURL
      },
      token
    })

    navigate('/')
  }

  const handleLogout = async () => {
    await signOut()
    clearAuthState()
    navigate('/login')
  }

  // If logged in, don't render anything (will redirect in useEffect)
  if (isLoggedIn) {
    return null
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