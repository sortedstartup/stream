import { atom, computed } from 'nanostores'
import { AddUserRequest, GetTenantsRequest, GetUsersRequest, TenantServiceClient, TenantUser, UserServiceClient } from '../proto/userservice'
import { CreateTenantRequest } from '../proto/userservice'
import { $authToken } from '../auth/store/auth'

// Tenant state atoms
export const $tenants = atom<TenantUser[]>([])
export const $currentTenant = atom<TenantUser | null>(null)
export const $isLoadingTenants = atom(false)
export const $tenantError = atom<string | null>(null)
export const $userTenantRoles = atom<Record<string, string>>({}) // tenantId -> role mapping

// Computed value to get current user's role in current tenant
export const $currentUserRole = computed([$currentTenant, $userTenantRoles], (currentTenant, userTenantRoles) => {
  if (!currentTenant) return null
  return userTenantRoles[currentTenant.tenant.id] || null
})

const apiUrl = import.meta.env.VITE_PUBLIC_API_URL?.replace(/\/$/, "")

// Initialize the UserService client with proper configuration
const userServiceClient = new UserServiceClient(
  apiUrl,
  {},
  {
    unaryInterceptors: [
      {
        intercept: (request, invoker) => {
          const metadata = request.getMetadata();
          metadata["authorization"] = $authToken.get();
          return invoker(request);
        },
      },
    ],
  }
)

const tenantServiceClient = new TenantServiceClient(
  apiUrl,
  {},
  {
    unaryInterceptors: [
      {
        intercept: (request, invoker) => {
          const metadata = request.getMetadata();
          metadata["authorization"] = $authToken.get();
          return invoker(request);
        },
      },
    ],
  }
)

// Tenant operations
export const loadUserTenants = async () => {
  try {
    $isLoadingTenants.set(true)
    $tenantError.set(null)
    
    const request = new GetTenantsRequest()
    const response = await userServiceClient.GetTenants(request, {})

    
    if (response.tenant_users) {
      $tenants.set(response.tenant_users)
      
      // Build role mapping from the response - now we have role information!
      const roleMapping: Record<string, string> = {}
      response.tenant_users.forEach(tenant => {
        roleMapping[tenant.tenant.id] = tenant.role.role
      })
      $userTenantRoles.set(roleMapping)
      
      // Set current tenant to personal tenant by default, or first tenant if no personal tenant
      const personalTenant = response.tenant_users.find(t => t.tenant.is_personal)
      const defaultTenant = personalTenant || response.tenant_users[0]
      if (defaultTenant) {
        $currentTenant.set(defaultTenant)
      }
    } else {
      $tenantError.set(response.message || 'Failed to load tenants')
    }
  } catch (error) {
    $tenantError.set('Failed to load tenants')
  } finally {
    $isLoadingTenants.set(false)
  }
}

export const createTenant = async (name: string, description: string = '') => {
  try {
    const request = new CreateTenantRequest({
      name,
      description
    })
    
    const response = await tenantServiceClient.CreateTenant(request, {})
    
    if (response.message && response.tenant_user) {
      const currentTenants = $tenants.get()
      $tenants.set([...currentTenants, response.tenant_user])
      
      // Update role mapping
      const currentRoles = $userTenantRoles.get()
      $userTenantRoles.set({
        ...currentRoles,
        [response.tenant_user.tenant.id]: response.tenant_user.role.role
      })
      
      return { success: true, tenant: response.tenant_user, error: null }
    } else {
      return { success: false, tenant: null, error: 'Failed to create workspace' }
    }
  } catch (error: any) {
    console.error('Error creating tenant:', error)
    
    // Extract meaningful error message from gRPC error
    let errorMessage = 'Failed to create workspace'
    if (error && error.message) {
      if (error.message.includes('already exists')) {
        errorMessage = 'A workspace with this name already exists'
      } else if (error.message.includes('InvalidArgument')) {
        errorMessage = 'Workspace name is required'
      } else {
        // Try to extract the actual error message from gRPC
        const match = error.message.match(/:\s*(.+)$/)
        if (match) {
          errorMessage = match[1]
        }
      }
    }
    
    return { success: false, tenant: null, error: errorMessage }
  }
}

export const switchTenant = (tenant: TenantUser) => {
  $currentTenant.set(tenant)
}

export const addUserToTenant = async (tenantId: string, username: string, role: string = 'member') => {
  try {
    const request = new AddUserRequest({
      tenant_id: tenantId,
      username: username,
      role
    })
    
    const response = await tenantServiceClient.AddUser(request, {})
    
    if (!response.message) {
      // Don't set global error - let modal handle it
      return false
    }
    
    return true
  } catch (error) {
    // Don't set global error - let modal handle it
    return false
  }
}

export const getTenantUsers = async (tenantId: string) => {
  try {
    $tenantError.set(null)
    
    const request = new GetUsersRequest({
      tenant_id: tenantId
    })
    
    const response = await tenantServiceClient.GetUsers(request, {})
    
    if (response.message && response.tenant_users) {
      return response.tenant_users
    } else {
      $tenantError.set(response.message || 'Failed to load tenant users')
      return []
    }
  } catch (error) {
    $tenantError.set('Failed to load tenant users')
    return []
  }
}

// Initialize tenants when auth state changes
$authToken.subscribe((token) => {
  if (token) {
    loadUserTenants()
  } else {
    $tenants.set([])
    $currentTenant.set(null)
  }
}) 