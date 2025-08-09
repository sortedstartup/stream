import React, { useState, useEffect } from 'react'
import { ToastContainer } from 'react-toastify'
import 'react-toastify/dist/ReactToastify.css'
import { Sidebar } from './Sidebar'
import { Header } from './Header'
import Footer from './Footer'
import { CreateWorkspaceModal } from '../modals'
import { UsageWarningBanner } from '../UsageWarningBanner'
import { createTenant, switchTenant } from '../../stores/tenants'
import { showSuccessToast } from '../../utils/toast'
import { useNavigate } from "react-router"
import { loadUserSubscription, $userSubscription } from '../../stores/payment'
import { $currentUser } from '../../auth/store/auth'
import { useStore } from '@nanostores/react'

export const Layout = ({ children }) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const navigate = useNavigate()
  const currentUser = useStore($currentUser)
  const subscription = useStore($userSubscription)

  useEffect(() => {
    const handler = () => setShowCreateModal(true)
    document.addEventListener('open-create-workspace', handler)
    
    return () => document.removeEventListener('open-create-workspace', handler)
  }, [])

  // Load subscription data when user is available
  useEffect(() => {
    if (currentUser?.uid) {
      loadUserSubscription()
    }
  }, [currentUser])

  const handleCreateTenant = async (name, description) => {
    const result = await createTenant(name, description || '')
    if (result.success) {
      setShowCreateModal(false)
      switchTenant(result.tenant)
      showSuccessToast('Workspace created successfully!')
      navigate('/workspace')
      return { success: true }
    }
    return { success: false, error: result.error }
  }

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
            {/* Global Usage Warning Banner */}
            {subscription && (
              <div className="mb-4">
                <UsageWarningBanner subscription={subscription} />
              </div>
            )}
            {children}
          </div>
        </main>
      </div>

      <footer className="bg-base-200 text-center py-4 md:ml-16">
        <Footer />
      </footer>

      {/* Global Create Workspace Modal */}
      <CreateWorkspaceModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSubmit={handleCreateTenant}
      />

      {/* Global Toast Container */}
      <ToastContainer
        position="top-center"
        autoClose={5000}
        hideProgressBar={false}
        newestOnTop={false}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover
        theme="light"
      />
    </div>
  )
}