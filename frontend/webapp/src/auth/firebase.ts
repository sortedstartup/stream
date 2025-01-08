// Import the functions you need from the SDKs you need
// compat packages are API compatible with namespaced code
import firebase from 'firebase/compat/app';
import * as firebaseui from 'firebaseui';
import 'firebaseui/dist/firebaseui.css';

// TODO: Add SDKs for Firebase products that you want to use
// https://firebase.google.com/docs/web/setup#available-libraries

// Your web app's Firebase configuration
// For Firebase JS SDK v7.20.0 and later, measurementId is optional
const firebaseConfig = {
  apiKey: "AIzaSyAhNUDCfFcQdhKQhNZWNTTtLz0fKdtumTE",
  authDomain: "prooftrack-7fbfc.firebaseapp.com",
  projectId: "prooftrack-7fbfc",
  storageBucket: "prooftrack-7fbfc.firebasestorage.app",
  messagingSenderId: "542619217550",
  appId: "1:542619217550:web:50d5ce7c928dd2a36ee3df",
  measurementId: "G-ZVT5CGR9TX"
};

// Initialize Firebase
//const app = initializeApp(firebaseConfig);
//const analytics = getAnalytics(app);

const fbapp = firebase.initializeApp(firebaseConfig)
const fbui = new firebaseui.auth.AuthUI(fbapp.auth());
export {fbapp, firebaseConfig, fbui}

//firebaseApp.initializeApp(firebaseConfig)

