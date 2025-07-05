import React from 'react'
import { useStore } from '@nanostores/react'
import { $tenants, $currentTenant, switchTenant } from '../stores/tenants'
import { TenantWithRole } from '../proto/userservice'
import { Briefcase, ChevronDown, User, Users, Check } from 'react-feather'

export const TenantSwitcher = () => {
  const tenants = useStore($tenants)
  const currentTenant = useStore($currentTenant)

  if (!tenants.length || !currentTenant) {
    return null
  }

  const handleTenantChange = (tenant: TenantWithRole) => {
    switchTenant(tenant)
  }

  return (
    <div className="dropdown dropdown-end">
      <div tabIndex={0} role="button" className="btn btn-ghost btn-sm flex items-center gap-2">
        <Briefcase className="w-4 h-4" />
        <span className="hidden sm:inline truncate max-w-24">
          {currentTenant.is_personal ? 'Personal' : currentTenant.name}
        </span>
        <ChevronDown className="w-3 h-3" />
      </div>
      <ul tabIndex={0} className="dropdown-content menu menu-sm z-[1] p-2 shadow bg-base-100 rounded-box w-52 mt-2">
        <li className="menu-title">Switch Workspace</li>
        {tenants.map((tenant) => (
          <li key={tenant.id}>
            <a 
              onClick={() => handleTenantChange(tenant)}
              className={`flex items-center gap-2 ${currentTenant.id === tenant.id ? 'active' : ''}`}
            >
              <div className="flex items-center gap-2 flex-1">
                {tenant.is_personal ? (
                  <User className="w-4 h-4" />
                ) : (
                  <Users className="w-4 h-4" />
                )}
                <span className="truncate">
                  {tenant.is_personal ? 'Personal' : tenant.name}
                </span>
              </div>
              {currentTenant.id === tenant.id && (
                <Check className="w-4 h-4" />
              )}
            </a>
          </li>
        ))}
      </ul>
    </div>
  )
} 