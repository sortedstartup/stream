import { atom } from 'nanostores'
import { auth } from '../providers/firebase-auth'
// Create atoms for auth state
export const $isLoggedIn = atom(false)
export const $currentUser = atom(null)
export const $authToken = atom("")

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

// Auth actions
export const setAuthState = ({ user, token }) => {
  $authToken.set(token)
  $currentUser.set(user)
  $isLoggedIn.set(true)
  localStorage.setItem('authToken', token)
}

export const clearAuthState = () => {
  $authToken.set("")
  $currentUser.set(null)
  $isLoggedIn.set(false)
  localStorage.removeItem('authToken')
} 