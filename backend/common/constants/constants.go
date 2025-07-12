package constants

// Channel roles
const (
	ChannelRoleOwner    = "owner"
	ChannelRoleUploader = "uploader"
	ChannelRoleViewer   = "viewer"
)

// Tenant roles
const (
	TenantRoleMember     = "member"
	TenantRoleSuperAdmin = "super_admin"
)

// ValidChannelRoles returns a slice of all valid channel roles
func ValidChannelRoles() []string {
	return []string{
		ChannelRoleOwner,
		ChannelRoleUploader,
		ChannelRoleViewer,
	}
}

// ValidTenantRoles returns a slice of all valid tenant roles
func ValidTenantRoles() []string {
	return []string{
		TenantRoleMember,
		TenantRoleSuperAdmin,
	}
}

// IsValidChannelRole checks if a role is a valid channel role
func IsValidChannelRole(role string) bool {
	validRoles := ValidChannelRoles()
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// IsValidTenantRole checks if a role is a valid tenant role
func IsValidTenantRole(role string) bool {
	validRoles := ValidTenantRoles()
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}
