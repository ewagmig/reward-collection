package admin

import (
	"context"
	"time"

	"github.com/op/go-logging"

	"github.com/awnumar/memguard"
	"github.com/spf13/viper"
)

const (
	realm               = "cubd"
	tokenHeader         = "Bearer"
	defaultTokenTimeout = time.Minute * time.Duration(30)
)

var adminLogger = logging.MustGetLogger("baas.admin")

var (
	securityKey = []byte{
		0x46, 0x68, 0x43, 0x6e, 0x77, 0x4f, 0x44, 0x74,
		0x62, 0x31, 0x52, 0x6a, 0x37, 0x73, 0x74, 0x79,
		0x75, 0x4f, 0x42, 0x6d, 0x47, 0x62, 0x74, 0x6a,
		0x69, 0x75, 0x63, 0x6d, 0x43, 0x43, 0x54, 0x42,
		0x6e, 0x62, 0x47, 0x6a, 0x70, 0x6e,
	}

	lockedSkBuffer *memguard.LockedBuffer
)

type AuthenticateUserFunc func(ctx context.Context, name, password, clientIP string) (*UserInfo, error)
type AuthorizateUserFunc func(ctx context.Context, userInfo *UserInfo, apiInfo *APIInfo) bool

func getTokeTimeout() time.Duration {
	t := viper.GetDuration("user.token.timeout")
	if t > 0 {
		return t
	}

	return defaultTokenTimeout
}

func utcNow() time.Time {
	return time.Now().UTC()
}

type UserInfo struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	Role          string   `json:"role"`
	IsActive      bool     `json:"is_active"`
	Blockchains   []uint   `json:"blockchains"`
	Orgs          []uint   `json:"orgs"`
	Permissions   []string `json:"permissions"`
	NotFirstLogin bool     `json:"not_first_login"`
}

func (ui *UserInfo) Reset() {
	ui.ID = 0
	ui.Name = ""
	ui.Role = ""
	ui.IsActive = false
	ui.Blockchains = nil
	ui.Orgs = nil
	ui.Permissions = nil
}

type APIInfo struct {
	Method       string   `json:"method"`
	Path         string   `json:"path"`
	AllowedRoles []string `json:"allowed_roles"`
}

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

//func UserInfoFromCtx(ctx context.Context) *UserInfo {
//	//adminLogger.Debugf("UserInfoFromCtx ctx.Value(utils.BAAS_CTX_USER_IDENTITY_KEY) is:", ctx.Value(utils.BAAS_CTX_USER_IDENTITY_KEY))
//	v := ctx.Value(utils.BAAS_CTX_USER_IDENTITY_KEY)
//	if v != nil {
//		return v.(*UserInfo)
//	}
//
//	return nil
//}

// CreateJWTFactory creates a JWT factory which contains auth middleware, login and token refreshing handler.
//func CreateJWTFactory(useQueryToken bool, authenFunc AuthenticateUserFunc, authorFunc AuthorizateUserFunc) *jwt.GinJWTMiddleware {
//	timeout := getTokeTimeout()
//	tokenLookup := "header:Authorization"
//	if useQueryToken {
//		tokenLookup = "query:token"
//	}
//
//
//	jwtMidd := &jwt.GinJWTMiddleware{
//		//Realm:       realm,
//		//Key:         lockedSkBuffer.Buffer(),
//		Timeout:     timeout,
//		MaxRefresh:  time.Hour * time.Duration(24),
//		IdentityKey: utils.BAAS_CTX_USER_IDENTITY_KEY,
//		PayloadFunc: func(data interface{}) jwt.MapClaims {
//			if v, ok := data.(*UserInfo); ok {
//				s, _ := json.Marshal(v)
//				return jwt.MapClaims{
//					utils.BAAS_CTX_USER_IDENTITY_KEY: string(s),
//				}
//			}
//			return jwt.MapClaims{}
//		},
//		IdentityHandler: func(c *gin.Context) interface{} {
//			return UserInfoFromCtx(c)
//		},
//
//		Authenticator: func(c *gin.Context) (interface{}, error) {
//			var loginVals login
//			if err := c.Bind(&loginVals); err != nil {
//				return "", jwt.ErrMissingLoginValues
//			}
//
//			data, err := authenFunc(c, loginVals.Username, loginVals.Password, c.ClientIP())
//			if err != nil {
//				return nil, err
//			}
//
//			// Put UserInfo into context, which will be used in LoginResponse
//			//adminLogger.Debugf("c.Set(utils.BAAS_CTX_USER_IDENTITY_KEY, data is:", data)
//			c.Set(utils.BAAS_CTX_USER_IDENTITY_KEY, data)
//			//adminLogger.Debugf("ctx.Value(utils.BAAS_CTX_USER_IDENTITY_KEY).(*UserInfo) is:", c.Value(utils.BAAS_CTX_USER_IDENTITY_KEY).(*UserInfo))
//
//			return data, nil
//
//		},
//		Authorizator: func(data interface{}, c *gin.Context) bool {
//			if data == nil {
//				return false
//			}
//
//			userInfo, ok := data.(*UserInfo)
//			if !ok {
//				return false
//			}
//
//			permInfo := c.MustGet(utils.BAAS_API_PERMISSION_KEY).(*utils.APIPermissionInfo)
//			if authorFunc != nil {
//				apiInfo := &APIInfo{
//					Method:       c.Request.Method,
//					Path:         permInfo.APIPath,
//					AllowedRoles: permInfo.AllowedRoles,
//				}
//				return authorFunc(c, userInfo, apiInfo)
//			}
//
//			return true
//		},
//		Unauthorized: func(c *gin.Context, code int, message string) {
//			switch code {
//			case http.StatusForbidden:
//				errors.ForbiddenError(errors.Forbidden, message).Write(c)
//			default:
//				errors.UnauthorizedError(errors.Unauthorized, message).Write(c)
//			}
//		},
//		LoginResponse: func(ctx *gin.Context, code int, token string, expire time.Time) {
//			userInfo := ctx.Value(utils.BAAS_CTX_USER_IDENTITY_KEY).(*UserInfo)
//			ctx.JSON(http.StatusOK, gin.H{
//				"code":          http.StatusOK,
//				"token":         token,
//				"expire":        expire.Format(time.RFC3339),
//				"id":            userInfo.ID,
//				"username":      userInfo.Name,
//				"userrole":      userInfo.Role,
//				"permissions":   userInfo.Permissions,
//				"blockchains":   userInfo.Blockchains,
//				"orgs":          userInfo.Orgs,
//				"notfirstlogin": userInfo.NotFirstLogin,
//			})
//		},
//
//		// TokenLookup is a string in the form of "<source>:<name>" that is used
//		// to extract token from the request.
//		// Optional. Default value "header:Authorization".
//		// Possible values:
//		// - "header:<name>"
//		// - "query:<name>"
//		// - "cookie:<name>"
//		TokenLookup: tokenLookup,
//		// TokenLookup: "query:token",
//		// TokenLookup: "cookie:token",
//
//		// TokenHeadName is a string in the header. Default value is "Bearer"
//		TokenHeadName: tokenHeader,
//
//		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
//		TimeFunc: utcNow,
//	}
//
//	if err := jwtMidd.MiddlewareInit(); err != nil {
//		panic(err)
//	}
//
//	return jwtMidd
//}
