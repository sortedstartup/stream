import React from 'react'
import { Sidebar } from './Sidebar'
import { Header } from './Header'

export const Layout = ({ children }) => {
  return (
    <div className="h-screen flex flex-col">
      <Header className="h-16" />
      <div className="flex flex-1">
        <Sidebar className="w-16" />
        <main className="flex-1 p-4 overflow-auto">
          {children}
        </main>
      </div>
    </div>
  )
}