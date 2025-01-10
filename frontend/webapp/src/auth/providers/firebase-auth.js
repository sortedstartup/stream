import { initializeApp } from 'firebase/app'
import { getAuth } from 'firebase/auth'
import * as firebaseui from 'firebaseui'
import 'firebaseui/dist/firebaseui.css'

// Your Firebase configuration
const firebaseConfig = {
    apiKey: "AIzaSyAdr9BQRPomQFz9r4lKvlnRSLGblAAW3ME",
    authDomain: "streams-testing-36d22.firebaseapp.com",
    projectId: "streams-testing-36d22",
    storageBucket: "streams-testing-36d22.firebasestorage.app",
    messagingSenderId: "981264158021",
    appId: "1:981264158021:web:91fc2ff96b62059f2f022c",
    measurementId: "G-SJTLFPX7EJ"
};

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
  signInFlow: 'popup',
  callbacks: {
    signInSuccessWithAuthResult: () => false // Don't redirect, we'll handle it
  }
}

// Initialize FirebaseUI
let ui
export const getUi = () => {
  if (!ui) {
    ui = new firebaseui.auth.AuthUI(auth)
  }
  return ui
}

export const startUi = (elementId, onSuccess) => {
  const ui = getUi()
  const config = {
    ...uiConfig,
    callbacks: {
      signInSuccessWithAuthResult: (authResult) => {
        onSuccess(authResult)
        return false
      }
    }
  }
  ui.start(elementId, config)
}


export const getCurrentUser = () => auth.currentUser

export const signOut = () => auth.signOut() 

