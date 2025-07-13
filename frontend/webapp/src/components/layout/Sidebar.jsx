import React from 'react'
import { Home, Settings, Users, Video, Circle, Upload as UploadIcon, Folder } from 'react-feather'
import { Link } from 'react-router'

export const Sidebar = ({ onItemClick }) => {
  const navItems = [
    { to: '/team', icon: Users, label: 'Team' },
    { to: '/', icon: Home, label: 'Home' },
    { to: '/channels', icon: Folder, label: 'Dashboard', className: 'text-blue-500' },
    { to: '/record', icon: Circle, label: 'Record', className: 'text-red-500' },
    { to: '/upload', icon: UploadIcon, label: 'Upload' },
    { to: '/settings', icon: Settings, label: 'Settings' },
  ]

  return (
    <nav className="h-full bg-base-200 flex flex-col w-20 md:w-20 border-r border-base-300">
      {/* Mobile Header */}
      <div className="md:hidden p-4 border-b border-base-300">
        <h2 className="text-lg font-semibold">Navigation</h2>
      </div>
      <div className="flex flex-col gap-2 p-2">
        {navItems.map(({ to, icon: Icon, label, className }) => (
          <Link
            key={to}
            to={to}
            className={
              `
                flex flex-col items-center justify-center p-2 hover:bg-base-300 rounded-lg transition-colors
                w-full
                md:flex-col md:items-center md:justify-center
                text-xs
              `
            }
            onClick={onItemClick}
          >
            <Icon size={22} className={className || ''} />
            <span className="mt-1">{label}</span>
          </Link>
        ))}
      </div>
    </nav>
  )
} 