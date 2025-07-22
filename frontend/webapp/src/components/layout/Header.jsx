import React from 'react'
import { useStore } from '@nanostores/react'
import { $currentUser, $isLoggedIn, clearAuthState } from '../../auth/store/auth'
import { useNavigate } from 'react-router'
import { TenantSwitcher } from '../TenantSwitcher'
import { Plus } from 'react-feather'

const toggleTheme = (e) => {
    const html = document.querySelector('html')
    if (e.target.checked) {
      html.setAttribute('data-theme', 'dark')
    } else {
      html.setAttribute('data-theme', 'light')
    }
}

export const Header = ({ onMenuClick }) => {
  const currentUser = useStore($currentUser)
  const isLoggedIn = useStore($isLoggedIn)
  const navigate = useNavigate()

  const handleOpenWorkspaceModal = () => {
    document.dispatchEvent(new CustomEvent("open-create-workspace"))
  }

  return (
    <div className="navbar bg-base-100 min-h-0 h-16 px-4">
      {/* Mobile Menu Button */}
      <div className="flex-none md:hidden">
        <button 
          className="btn btn-square btn-ghost"
          onClick={onMenuClick}
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="inline-block w-5 h-5 stroke-current">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 6h16M4 12h16M4 18h16"></path>
          </svg>
        </button>
      </div>

      <div className="flex-1">
        <a className="btn btn-ghost text-xl p-0">Stream</a>
      </div>

      <div className="flex-none flex items-center gap-4">
        {isLoggedIn && <TenantSwitcher />}
         {isLoggedIn && (
          <button
            className="btn btn-primary btn-sm"
            onClick={handleOpenWorkspaceModal}
          >
            <Plus className="w-4 h-4" />
            New Workspace
          </button>
        )}
        <ThemeSelector />
        {isLoggedIn && <UserMenu />}
      </div>
    </div>
  )
}

export const ThemeSelector = () => {
    return (
      <label className="grid cursor-pointer place-items-center">
  <input
    type="checkbox"
    value="synthwave"
    onChange={toggleTheme}
    className="toggle theme-controller bg-base-content col-span-2 col-start-1 row-start-1" />
  <svg
    className="stroke-base-100 fill-base-100 col-start-1 row-start-1"
    xmlns="http://www.w3.org/2000/svg"
    width="14"
    height="14"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round">
    <circle cx="12" cy="12" r="5" />
    <path
      d="M12 1v2M12 21v2M4.2 4.2l1.4 1.4M18.4 18.4l1.4 1.4M1 12h2M21 12h2M4.2 19.8l1.4-1.4M18.4 5.6l1.4-1.4" />
  </svg>
  <svg
    className="stroke-base-100 fill-base-100 col-start-2 row-start-1"
    xmlns="http://www.w3.org/2000/svg"
    width="14"
    height="14"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round">
    <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
  </svg>
</label>
    )
}

export const UserMenu = () => {

  const currentUser = useStore($currentUser)
  const navigate = useNavigate()

  const handleLogout = async () => {
    clearAuthState()
    navigate('/login')
  }

  return (
    <div className="dropdown dropdown-end relative">
      <div tabIndex={0} role="button" className="btn btn-ghost btn-sm p-0">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="w-5 h-5 stroke-current">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"></path>
          <circle cx="12" cy="7" r="4"></circle>
        </svg>
      </div>
      <ul tabIndex={0} className="dropdown-content menu menu-sm z-[1] p-2 shadow bg-base-100 rounded-box w-52 mt-2">
        <li className="menu-title">{currentUser?.displayName || 'User'}</li>
        <li><a href="/profile">Profile</a></li>
        <li><a onClick={handleLogout}>Logout</a></li>
      </ul>
    </div>
  )
}
