import { atom, computed } from 'nanostores'
import { UserServiceClient } from '../proto/userservice'
import { CreateTenantRequest, GetUserTenantsRequest, AddUserToTenantRequest, GetTenantUsersRequest, TenantWithRole } from '../proto/userservice'
import { $authToken } from '../auth/store/auth'

// Tenant state atoms
export const $tenants = atom<TenantWithRole[]>([])
export const $currentTenant = atom<TenantWithRole | null>(null)
export const $isLoadingTenants = atom(false)
export const $tenantError = atom<string | null>(null)
export const $userTenantRoles = atom<Record<string, string>>({}) // tenantId -> role mapping

// Computed value to get current user's role in current tenant
export const $currentUserRole = computed([$currentTenant, $userTenantRoles], (currentTenant, userTenantRoles) => {
  if (!currentTenant) return null
  return userTenantRoles[currentTenant.id] || null
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


// Tenant operations
export const loadUserTenants = async () => {
  try {
    $isLoadingTenants.set(true)
    $tenantError.set(null)
    
    const request = new GetUserTenantsRequest()
    const response = await userServiceClient.GetUserTenants(request, {})

    
    if (response.success) {
      $tenants.set(response.tenants)
      
      // Build role mapping from the response - now we have role information!
      const roleMapping: Record<string, string> = {}
      response.tenants.forEach(tenant => {
        roleMapping[tenant.id] = tenant.role
      })
      $userTenantRoles.set(roleMapping)
      
      // Set current tenant to personal tenant by default, or first tenant if no personal tenant
      const personalTenant = response.tenants.find(t => t.is_personal)
      const defaultTenant = personalTenant || response.tenants[0]
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
    
    const response = await userServiceClient.CreateTenant(request, {})
    
    if (response.success && response.tenant) {
      // Add the new tenant to the list with super_admin role (creator gets super_admin)
      const newTenantWithRole = new TenantWithRole({
        id: response.tenant.id,
        name: response.tenant.name,
        description: response.tenant.description,
        is_personal: response.tenant.is_personal,
        created_at: response.tenant.created_at,
        created_by: response.tenant.created_by,
        role: 'super_admin' // Creator always gets super_admin role
      })
      
      const currentTenants = $tenants.get()
      $tenants.set([...currentTenants, newTenantWithRole])
      
      // Update role mapping
      const currentRoles = $userTenantRoles.get()
      $userTenantRoles.set({
        ...currentRoles,
        [newTenantWithRole.id]: 'super_admin'
      })
      
      return newTenantWithRole
    } else {
      // Don't set global error - let modal handle it
      return null
    }
  } catch (error) {
    // Don't set global error - let modal handle it
    return null
  }
}

export const switchTenant = (tenant: TenantWithRole) => {
  $currentTenant.set(tenant)
}

export const addUserToTenant = async (tenantId: string, username: string, role: string = 'member') => {
  try {
    const request = new AddUserToTenantRequest({
      tenant_id: tenantId,
      username: username,
      role
    })
    
    const response = await userServiceClient.AddUserToTenant(request, {})
    
    if (!response.success) {
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
    
    const request = new GetTenantUsersRequest({
      tenant_id: tenantId
    })
    
    const response = await userServiceClient.GetTenantUsers(request, {})
    
    if (response.success) {
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