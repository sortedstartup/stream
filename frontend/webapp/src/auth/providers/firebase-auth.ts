import { initializeApp } from 'firebase/app'
import { getAuth, User, AuthError } from 'firebase/auth'
import * as firebaseui from 'firebaseui'
import 'firebaseui/dist/firebaseui.css'

// Your Firebase configuration
const firebaseConfig = JSON.parse(import.meta.env.VITE_FIREBASE_CONFIG);

// Initialize Firebase
export const app = initializeApp(firebaseConfig)
export const auth = getAuth(app)

// Configure FirebaseUI
const uiConfig = {
  signInOptions: [
    // Add sign-in methods you want to support
    'google.com',
    'facebook.com',
    'email'
  ],
  signInFlow: 'popup' as const,
  callbacks: {
    signInSuccessWithAuthResult: () => false // Don't redirect, we'll handle it
  }
}

// Initialize FirebaseUI
let ui: firebaseui.auth.AuthUI | undefined
export const getUi = () => {
  if (!ui) {
    ui = new firebaseui.auth.AuthUI(auth)
  }
  return ui
}

interface AuthResult {
  user: User
  credential: any
  additionalUserInfo?: any
}

export const startUi = (elementId: string, onSuccess: (result: AuthResult) => void) => {
  const ui = getUi()
  const config = {
    ...uiConfig,
    callbacks: {
      signInSuccessWithAuthResult: (authResult: AuthResult) => {
        onSuccess(authResult)
        return false
      }
    }
  }
  ui.start(elementId, config)
}

export const logout = async (): Promise<boolean> => {
  try {
    await auth.signOut()
    return true
  } catch (error) {
    console.error('Error signing out:', error)
    return false
  }
}

export const getCurrentUser = (): User | null => auth.currentUser

export const signOut = () => auth.signOut() 

