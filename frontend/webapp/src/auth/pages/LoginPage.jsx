import { useEffect, useRef } from 'react'
import { useNavigate } from 'react-router'
import { useStore } from '@nanostores/react'
import { $isLoggedIn } from '../store/auth'
import { startUi, signOut } from '../providers/firebase-auth'
import { setAuthState, clearAuthState } from '../store/auth'

export default function LoginPage() {
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

  if (isLoggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="max-w-md w-full p-6 bg-white rounded-lg shadow-lg text-center">
          <h2 className="text-2xl font-bold mb-6">You are logged in</h2>
          <button
            onClick={handleLogout}
            className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"
          >
            Logout
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="max-w-md w-full p-6 bg-white rounded-lg shadow-lg">
        <h2 className="text-2xl font-bold mb-6">Login</h2>
        <div id="firebaseui-auth-container" ref={firebaseUiContainerRef}></div>
      </div>
    </div>
  )
} 