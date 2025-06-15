import React, { useState } from 'react'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import Footer from './Footer'

export const Layout = ({ children }) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false)

  return (
    <div className="flex flex-col min-h-screen">
      <header className="fixed top-0 left-0 right-0 z-50 bg-base-100 shadow-sm h-16">
        <Header onMenuClick={() => setIsSidebarOpen(!isSidebarOpen)} />
      </header>

      <div className="flex flex-1 pt-16"> 
        {/* Mobile Sidebar Overlay */}
        {isSidebarOpen && (
          <div 
            className="fixed inset-0 bg-black bg-opacity-50 z-40 md:hidden"
            onClick={() => setIsSidebarOpen(false)}
          />
        )}

        {/* Sidebar */}
        <aside className={`
          fixed md:static
          top-16 bottom-0
          w-64 md:w-16
          z-40 bg-base-100 border-r
          transform transition-transform duration-200 ease-in-out
          ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'}
        `}>
          <Sidebar onItemClick={() => setIsSidebarOpen(false)} />
        </aside>

        {/* Main Content */}
        <main className="flex-1 w-full md:ml-16 px-4 overflow-x-hidden">
          <div className="max-w-screen-xl mx-auto py-4">
            {children}
          </div>
        </main>
      </div>

      <footer className="bg-base-200 text-center py-4 md:ml-16">
        <Footer />
      </footer>
    </div>
  )
}