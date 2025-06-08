import React from 'react'
import { Home, Settings, Users, Video, Circle, Folder, Upload } from 'react-feather'
import { Link } from 'react-router'

export const Sidebar = ({ className }) => {
  return (
    <nav className={`bg-base-200 flex flex-col ${className}`}>
      {/* Section 1 - Team */}
      <div className="flex flex-col gap-2 p-2 border-b border-base-300">
        <Link 
          to="/team"
          className="flex flex-col items-center p-2 hover:bg-base-300 rounded-lg"
        >
          <Users size={20} />
          <span className="text-xs">Team</span>
        </Link>
      </div>

      {/* Section 2 - Main Navigation */}
      <div className="flex flex-col gap-2 p-2">
        <Link 
          to="/"
          className="active flex flex-col items-center p-2 hover:bg-base-300 rounded-lg"
        >
          <Home size={20} />
          <span className="text-xs">Home</span>
        </Link>

        <Link 
          to="/record"
          className="flex flex-col items-center p-2 hover:bg-base-300 rounded-lg"
        >
          <Circle size={20} className="text-red-500" />
          <span className="text-xs">Record</span>
        </Link>

        <Link 
          to="/upload"
          className="flex flex-col items-center p-2 hover:bg-base-300 rounded-lg"
        >
          <Upload size={20} className="text-blue-500" />
          <span className="text-xs">Upload</span>
        </Link>

        <Link 
          to="/videos"
          className="flex flex-col items-center p-2 hover:bg-base-300 rounded-lg"
        >
          <Video size={20} />
          <span className="text-xs">Videos</span>
        </Link>

        <Link 
          to="/spaces"
          className="flex flex-col items-center p-2 hover:bg-base-300 rounded-lg"
        >
          <Folder size={20} />
          <span className="text-xs">Spaces</span>
        </Link>

        <Link 
          to="/settings"
          className="flex flex-col items-center p-2 hover:bg-base-300 rounded-lg"
        >
          <Settings size={20} />
          <span className="text-xs">Settings</span>
        </Link>
      </div>
    </nav>
  )
} 