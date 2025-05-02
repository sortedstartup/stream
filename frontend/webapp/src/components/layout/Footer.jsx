import React from 'react'
import { useStore } from '@nanostores/react'
import { $currentUser } from '../../auth/store/auth'

export const Footer = () => {
  const currentUser = useStore($currentUser)
  
  return (
    <footer className="footer footer-center p-4 bg-base-200 text-base-content">
      <aside>
        <p>&copy; {new Date().getFullYear()} Stream - All rights reserved</p>
        {currentUser && (
          <p className="text-xs opacity-70 mt-1">
            Logged in as {currentUser.email}
          </p>
        )}
      </aside>
    </footer>
  )
}

export default Footer