import React from 'react'
import { Link, useLocation } from 'react-router-dom'

export const Breadcrumb = () => {
  const location = useLocation()
  const pathParts = location.pathname.split('/').filter(Boolean)

  return (
    <div className="text-sm breadcrumbs mb-4">
      <ul>
        <li><Link to="/">Home</Link></li>
        {pathParts.map((part, index) => {
          const path = '/' + pathParts.slice(0, index + 1).join('/')
          return (
            <li key={path}>
              <Link to={path}>{decodeURIComponent(part)}</Link>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
