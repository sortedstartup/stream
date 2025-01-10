import { atom } from 'nanostores'

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