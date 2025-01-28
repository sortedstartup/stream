import React from 'react'
import { useStore } from '@nanostores/react'
import { Navigate } from 'react-router'
import { $isLoggedIn } from '../store/auth'

export default function ProtectedRoute({ children }) {
  const isLoggedIn = useStore($isLoggedIn)

  if (!isLoggedIn) {
    return <Navigate to="/login" replace />
  }

  return children
} 