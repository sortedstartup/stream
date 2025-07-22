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
  
  const [showAddUserModal, setShowAddUserModal] = useState(false)
  const [tenantUsers, setTenantUsers] = useState([])
  const [loadingUsers, setLoadingUsers] = useState(false)

  // Check if user can view/manage members (super_admin only)
  const canManageMembers = currentUserRole === 'super_admin'
  
  // Load tenant users when current tenant changes
  useEffect(() => {
    if (!currentTenant || currentTenant.tenant.is_personal || !canManageMembers) {
      setTenantUsers([])
      return
    }
    
    const loadUsers = async () => {
      setLoadingUsers(true)
      try {
        const users = await getTenantUsers(currentTenant.tenant.id)
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

  const handleAddUser = async (username, role) => {
    if (username && currentTenant && currentTenant.tenant) {
      const success = await addUserToTenant(currentTenant.tenant.id, username, role)
      if (success) {
        setShowAddUserModal(false)
        // Refresh the user list
        const users = await getTenantUsers(currentTenant.tenant.id)
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
        {currentTenant && currentTenant.tenant &&  (
          <div className="card bg-base-200">
            <div className="card-body">
              <div className="flex justify-between items-start">
                <div>
                  <h2 className="card-title flex items-center gap-2">
                    {currentTenant.tenant.is_personal ? (
                      <User className="w-6 h-6" />
                    ) : (
                      <Users className="w-6 h-6" />
                    )}
                    {currentTenant.tenant.name}
                    <span className={`badge ${currentTenant.tenant.is_personal ? 'badge-info' : 'badge-success'}`}>
                      {currentTenant.tenant.is_personal ? 'Personal' : 'Organization'}
                    </span>
                  </h2>
                  {currentTenant.tenant.description && (
                    <p className="text-base-content/70 mt-2">{currentTenant.tenant.description}</p>
                  )}
                </div>
                {!currentTenant.tenant.is_personal && canManageMembers && (
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
        {canManageMembers && currentTenant && currentTenant.tenant && !currentTenant.tenant.is_personal && (
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
                      </tr>
                    </thead>
                    <tbody>
                      {tenantUsers && tenantUsers.length > 0 && tenantUsers.map((tenantUser) => (
                        <tr>
                          <td>
                            <div className="flex items-center gap-3">
                              <div>
                                <div className="font-medium">{tenantUser.user?.username || 'Unknown'}</div>
                                <div className="text-sm text-gray-500">{tenantUser.user?.email || 'No email'}</div>
                              </div>
                            </div>
                          </td>
                          <td>
                            <span className={`badge ${tenantUser.role?.role === 'super_admin' ? 'badge-primary' : 'badge-secondary'}`}>
                              {tenantUser.role?.role}
                              </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Modals */}
      <AddUserModal
        isOpen={showAddUserModal}
        onClose={() => setShowAddUserModal(false)}
        onSubmit={handleAddUser}
      />
    </Layout>
  )
}