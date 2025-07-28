import React from 'react'
import { useStore } from '@nanostores/react'
import { useNavigate, useLocation } from 'react-router'
import { $tenants, $currentTenant, switchTenant } from '../stores/tenants'
import { Briefcase, ChevronDown, User, Users, Check } from 'react-feather'
import { TenantUser } from '../proto/userservice'

export const TenantSwitcher = () => {
  const tenants = useStore($tenants)
  const currentTenant = useStore($currentTenant)
  const navigate = useNavigate()
  const location = useLocation()

  if (!tenants.length || !currentTenant) {
    return null 
  }

  const handleTenantChange = async (tenant: TenantUser) => {
    if (tenant.tenant.id === currentTenant.tenant.id) return;

    await switchTenant(tenant)

    const path = location.pathname

    // Match paths like /channel/abc123, /video/xyz, /record/123
    const isTenantScopedResourcePath = /^\/[^/]+\/[^/]+$/.test(path)

    if (isTenantScopedResourcePath) {
      navigate('/channels', { replace: true })
    } else {
      navigate(path, { replace: true })
    }
  }

  return (
    <div className="dropdown dropdown-end">
      <div tabIndex={0} role="button" className="btn btn-ghost btn-sm flex items-center gap-2">
        <Briefcase className="w-4 h-4" />
        <span className="hidden sm:inline truncate max-w-24">
          {currentTenant.tenant.is_personal ? 'Personal' : currentTenant.tenant.name}
        </span>
        <ChevronDown className="w-3 h-3" />
      </div>
      <ul tabIndex={0} className="dropdown-content menu menu-sm z-[1] p-2 shadow bg-base-100 rounded-box w-52 mt-2">
        <li className="menu-title">Switch Workspace</li>
        {tenants.map((tenant) => (
          <li key={tenant.tenant.id}>
            <a 
              onClick={() => handleTenantChange(tenant)}
              className={`flex items-center gap-2 ${currentTenant.tenant.id === tenant.tenant.id ? 'active' : ''}`}
            >
              <div className="flex items-center gap-2 flex-1">
                {tenant.tenant.is_personal ? (
                  <User className="w-4 h-4" />
                ) : (
                  <Users className="w-4 h-4" />
                )}
                <span className="truncate">
                  {tenant.tenant.is_personal ? 'Personal' : tenant.tenant.name}
                </span>
              </div>
              {currentTenant.tenant.id === tenant.tenant.id && (
                <Check className="w-4 h-4" />
              )}
            </a>
          </li>
        ))}
      </ul>
    </div>
  )
} 