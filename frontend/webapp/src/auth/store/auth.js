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
  // Optionally: Validate token here or fetch user data
}

// Auth actions
export const login = async (credentials) => {
  try {
    // TODO: Replace with actual API call
    const response = await mockLogin(credentials)
    
    // Store token and user data
    const { user, token } = response
    $authToken.set(token)
    $currentUser.set(user)
    $isLoggedIn.set(true)
    
    // Persist token
    localStorage.setItem('authToken', token)
    
    return { success: true }
  } catch (error) {
    return { success: false, error: error.message }
  }
}

export const logout = () => {
  // Clear all auth state
  $authToken.set(null)
  $currentUser.set(null)
  $isLoggedIn.set(false)
  
  // Remove from localStorage
  localStorage.removeItem('authToken')
}

// Mock function - replace with actual API call
const mockLogin = async ({ email, password }) => {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 500))
  
  if (email === 'test@example.com' && password === 'password') {
    return {
      user: {
        id: 1,
        email: 'test@example.com',
        name: 'Test User'
      },
      token: 'mock-jwt-token-' + Math.random()
    }
  }
  throw new Error('Invalid credentials')
} 