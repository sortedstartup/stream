import { atom } from 'nanostores'
import { auth, logout } from '../providers/firebase-auth'
import { User } from 'firebase/auth'

// Create atoms for auth state
export const $isLoggedIn = atom(false)
export const $currentUser = atom<User | null>(null)
export const $authToken = atom<string>("")

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

initTokenRefreshHandler()

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