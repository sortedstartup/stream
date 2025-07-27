import { atom } from 'nanostores'
import { auth, logout } from '../providers/firebase-auth'
import { User } from 'firebase/auth'

// Create atoms for auth state
export const $isLoggedIn = atom(false)
export const $currentUser = atom<User | null>(null)
export const $authToken = atom<string>("")
export const $authInitialized = atom(false)

// Initialize auth state from localStorage
const savedToken = localStorage.getItem('authToken')
if (savedToken) {
  $authToken.set(savedToken)
  $isLoggedIn.set(true)
}

export const initTokenRefreshHandler = () => {
  return auth.onIdTokenChanged(async (user) => {
    if (user) {
      const token = await user.getIdToken()
      localStorage.setItem('authToken', token)
    } else {
      localStorage.removeItem('authToken') 
    }
  })
}

export const initAuthStateHandler = () => {
  return auth.onAuthStateChanged(async (user) => {
    if (user) {
      const token = await user.getIdToken()
      setAuthState({ user, token })
    } else {
      clearAuthState()
    }
    // Set auth as initialized once we've processed the initial auth state
    $authInitialized.set(true)
  })
}

initTokenRefreshHandler()
initAuthStateHandler()

interface AuthState {
  user: User | null
  token: string
}

// Auth actions
export const setAuthState = ({ user, token }: AuthState) => {
  $authToken.set(token)
  $currentUser.set(user)
  $isLoggedIn.set(true)
  localStorage.setItem('authToken', token)
}

export const clearAuthState = () => {
  logout()
  $authToken.set("")
  $currentUser.set(null)
  $isLoggedIn.set(false)
  localStorage.removeItem('authToken')
} 