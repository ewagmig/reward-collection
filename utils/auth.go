package utils

// AuthType is the enum for authentication types: basic and token
type AuthType int

const (
	TokenAuth AuthType = iota
	BasicAuth
	NoAuth
)

const (
	UserRoleRoot          = "root"   //baas admin
	UserRoleTenant        = "tenant" //tenant
	UserRoleAudit         = "audit"  //network audit
	UserRoleNetworkAdmin  = "admin"  //network admin
	UserRoleNetworkNormal = "normal" //network normal user
	AllPermissionName     = "all"
)

type APIPermissionInfo struct {
	APIPath      string
	AllowedRoles []string
	Permissions  []string
}

var (
	AllRoles              = []string{UserRoleRoot, UserRoleTenant, UserRoleAudit, UserRoleNetworkAdmin, UserRoleNetworkNormal}
	NotAuditRoles         = []string{UserRoleRoot, UserRoleTenant, UserRoleNetworkAdmin, UserRoleNetworkNormal}
	OnlyRootRole          = []string{UserRoleRoot}
	OnlyRoleTenant        = []string{UserRoleTenant}
	CreateUserRole        = []string{UserRoleRoot, UserRoleTenant, UserRoleNetworkAdmin}
	OnlyNetworkAdminRole  = []string{UserRoleNetworkAdmin}
	OnlyNetworkNormalRole = []string{UserRoleNetworkNormal}
	OnlyAuditRole         = []string{UserRoleAudit}
	AllPermissions        = []string{AllPermissionName}
)
