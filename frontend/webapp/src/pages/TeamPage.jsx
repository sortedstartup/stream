import React, { useState, useEffect } from 'react'
import { useStore } from '@nanostores/react'
import { Layout } from "../components/layout/Layout"
import { $tenants, $currentTenant, $tenantError, createTenant, getTenantUsers, addUserToTenant, $currentUserRole } from '../stores/tenants'
import { Plus, AlertCircle, User, Users, UserPlus, Check, X } from 'react-feather'
import { CreateWorkspaceModal, AddUserModal } from '../components/modals'

export const TeamPage = () => {
  const tenants = useStore($tenants)
  const currentTenant = useStore($currentTenant)
  const tenantError = useStore($tenantError)
  const currentUserRole = useStore($currentUserRole)
  
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showAddUserModal, setShowAddUserModal] = useState(false)
  const [tenantUsers, setTenantUsers] = useState([])
  const [loadingUsers, setLoadingUsers] = useState(false)
  
  // Check if user can view/manage members (super_admin only)
  const canManageMembers = currentUserRole === 'super_admin'
  
  // Load tenant users when current tenant changes
  useEffect(() => {
    if (!currentTenant || currentTenant.is_personal || !canManageMembers) {
      setTenantUsers([])
      return
    }
    
    const loadUsers = async () => {
      setLoadingUsers(true)
      try {
        const users = await getTenantUsers(currentTenant.id)
        setTenantUsers(users)
      } catch (error) {
        console.error('Error loading tenant users:', error)
        setTenantUsers([])
      } finally {
        setLoadingUsers(false)
      }
    }

    loadUsers()
  }, [currentTenant, canManageMembers])

  const handleCreateTenant = async (name, description) => {
    const newTenant = await createTenant(name, description || '')
    if (newTenant) {
      setShowCreateModal(false)
      return true
    }
    return false
  }

  const handleAddUser = async (username, role) => {
    if (username && currentTenant) {
      const success = await addUserToTenant(currentTenant.id, username, role)
      if (success) {
        setShowAddUserModal(false)
        // Refresh the user list
        const users = await getTenantUsers(currentTenant.id)
        setTenantUsers(users)
        return true
      }
    }
    return false
  }

  const dismissError = () => {
    // Clear the error from the store
    $tenantError.set(null)
  }

  return (
    <Layout>
      <div className="space-y-8">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold">Team Management</h1>
            <p className="text-base-content/70 mt-2">
              Manage your workspaces and team members
            </p>
          </div>
          <button 
            className="btn btn-primary"
            onClick={() => setShowCreateModal(true)}
          >
            <Plus className="w-5 h-5" />
            New Workspace
          </button>
        </div>

        {/* Error Alert */}
        {tenantError && (
          <div className="alert alert-error">
            <AlertCircle className="stroke-current shrink-0 h-6 w-6" />
            <span>{tenantError}</span>
            <button 
              className="btn btn-sm btn-ghost btn-circle ml-auto"
              onClick={dismissError}
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        )}

        {/* Current Tenant Info */}
        {currentTenant && (
          <div className="card bg-base-200">
            <div className="card-body">
              <div className="flex justify-between items-start">
                <div>
                  <h2 className="card-title flex items-center gap-2">
                    {currentTenant.is_personal ? (
                      <User className="w-6 h-6" />
                    ) : (
                      <Users className="w-6 h-6" />
                    )}
                    {currentTenant.name}
                    <span className={`badge ${currentTenant.is_personal ? 'badge-info' : 'badge-success'}`}>
                      {currentTenant.is_personal ? 'Personal' : 'Organization'}
                    </span>
                  </h2>
                  {currentTenant.description && (
                    <p className="text-base-content/70 mt-2">{currentTenant.description}</p>
                  )}
                </div>
                {!currentTenant.is_personal && canManageMembers && (
                  <button 
                    className="btn btn-outline btn-sm"
                    onClick={() => setShowAddUserModal(true)}
                  >
                    <UserPlus className="w-4 h-4" />
                    Add User
                  </button>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Team Members - Only show for super admins */}
        {canManageMembers && (
          <div className="card bg-base-100">
            <div className="card-body">
              <h2 className="card-title flex items-center gap-2">
                <Users className="w-5 h-5" />
                Team Members
              </h2>
              
              {loadingUsers ? (
                <div className="flex justify-center py-8">
                  <span className="loading loading-spinner loading-md"></span>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="table">
                    <thead>
                      <tr>
                        <th>Member</th>
                        <th>Role</th>
                        <th>Joined</th>
                      </tr>
                    </thead>
                    <tbody>
                      {tenantUsers.map((tenantUser) => (
                        <tr key={tenantUser.id}>
                          <td>
                            <div className="flex items-center gap-3">
                              <div className="avatar placeholder">
                                <div className="bg-neutral text-neutral-content w-8 h-8 rounded-full flex items-center justify-center leading-none">
                                  <span className="text-xs">
                                    {tenantUser.user?.username?.charAt(0)?.toUpperCase() || '?'}
                                  </span>
                                </div>
                              </div>
                              <div>
                                <div className="font-medium">{tenantUser.user?.username || 'Unknown'}</div>
                                <div className="text-sm text-gray-500">{tenantUser.user?.email || 'No email'}</div>
                              </div>
                            </div>
                          </td>
                          <td>
                            <span className={`badge ${tenantUser.role === 'super_admin' ? 'badge-primary' : 'badge-secondary'}`}>
                              {tenantUser.role}
                            </span>
                          </td>
                          <td>{new Date(tenantUser.created_at?.seconds * 1000).toLocaleDateString()}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        )}

        {/* All Workspaces */}
        <div className="card bg-base-100">
          <div className="card-body">
            <h3 className="card-title">All Workspaces</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {tenants.map((tenant) => (
                <div key={tenant.id} className="card bg-base-200 shadow-sm">
                  <div className="card-body p-4">
                    <h4 className="card-title text-lg flex items-center gap-2">
                      {tenant.is_personal ? (
                        <User className="w-5 h-5" />
                      ) : (
                        <Users className="w-5 h-5" />
                      )}
                      <span className="truncate">{tenant.name}</span>
                    </h4>
                    {tenant.description && (
                      <p className="text-sm text-base-content/70">{tenant.description}</p>
                    )}
                    <div className="card-actions justify-end mt-2">
                      <span className={`badge ${tenant.is_personal ? 'badge-info' : 'badge-success'}`}>
                        {tenant.is_personal ? 'Personal' : 'Organization'}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Modals */}
      <CreateWorkspaceModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSubmit={handleCreateTenant}
      />
      
      <AddUserModal
        isOpen={showAddUserModal}
        onClose={() => setShowAddUserModal(false)}
        onSubmit={handleAddUser}
      />
    </Layout>
  )
}