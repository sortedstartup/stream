import React from 'react'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import Footer from './Footer'

export const Layout = ({ children }) => {
  return (
    <div className="flex flex-col min-h-screen">
      <header className="fixed top-0 left-0 right-0 z-50 bg-base-100 shadow-sm h-16">
        <Header />
      </header>

      <div className="flex flex-1 pt-16"> 
        <aside className="fixed left-0 top-16 bottom-0 w-16 z-40 bg-base-100 border-r">
          <Sidebar />
        </aside>

        <main className="flex-1 ml-56 pl-4 pr-4 overflow-x-hidden">
          <div className="max-w-screen-xl mx-auto">
            {children}
          </div>
        </main>
      </div>

      <footer className="bg-base-200 text-center py-4 ml-56">
        <Footer />
      </footer>
    </div>
  )
}