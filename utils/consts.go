package utils

import "github.com/gin-gonic/gin"

// HandlerFunc is wrap function
type HandlerFunc func(*gin.Context)

const (
	BAAS_DB_BLOCKCHAINS_KEY    = "_BAAS_DB_BLOCKCHIAN_KEY_"
	BAAS_CTX_RESPONSE_KEY      = "_BAAS_CTX_RESPONSE_KEY_"
	BAAS_CTX_AUDIT_LOG_KEY     = "_BAAS_CTX_AUDIT_LOG_KEY_"
	BAAS_CTX_USER_IDENTITY_KEY = "_BAAS_CTX_USER_IDENTITY_KEY_"
	BAAS_API_AUDIT_INFO_KEY    = "_BAAS_API_AUDIT_INFO_KEY_"
	BAAS_API_AUTH_TYPE_KEY     = "_BAAS_API_AUTH_TYPE_"
	BAAS_API_PERMISSION_KEY    = "_BAAS_API_PERMISSION_TYPE_"
	NODE_ID_NAME_KEY           = "_NODE_ID_NAME_"
	DEFAULT_ESCC               = "escc"
	DEFAULT_VSCC               = "vscc"
)
