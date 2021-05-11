package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ewagmig/rewards-collection/utils"
	"github.com/gin-gonic/gin"
	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

const (
	defaultRequestTimeout = time.Duration(600) * time.Second
	//defaultListenAddr     = ":8080"
	defaultVersion = "v1"
)

var logger = logging.MustGetLogger("common.server")

// Mode represents DEV or PROD
type Mode int

const (
	// DEV is for develop.
	DEV Mode = iota
	// PROD is for production.
	PROD
)

// Server is a API server to servcer http requests.
type Server struct {
	opts            options
	server          *http.Server
	controllerList  []Controller
	//middlewaresList []Middleware
	stopChan        chan struct{}
}

type options struct {
	apiVersion       string
	mode             Mode
	basePath         string
	controllerFilter func(Controller) bool
	//middlewareFilter func(Middleware) bool
	requestTimeout   time.Duration
}

var defaultOptions = options{
	basePath:         "/api",
	controllerFilter: func(Controller) bool { return false },
	//middlewareFilter: func(Middleware) bool { return false },
	requestTimeout:   defaultRequestTimeout,
}

// Option sets server options such as version, baseURL and controllers, etc.
type Option func(*options)

// WithVersion returns an Option that sets the api version.
func WithVersion(v string) Option {
	return func(o *options) {
		o.apiVersion = v
	}
}

// WithMode returns an Option that sets the server mode, DEV or PROD.
func WithMode(m Mode) Option {
	return func(o *options) {
		o.mode = m
	}
}

// RequestBasePath returns an Option that sets the base releative path.
func RequestBasePath(u string) Option {
	return func(o *options) {
		if !strings.HasPrefix(u, "/") {
			logger.Panic("Base Path should begin with /")
		}
		o.basePath = u
	}
}

// ControllerFilter returns an Option that sets the controller filter.
func ControllerFilter(f func(Controller) bool) Option {
	return func(o *options) {
		if f != nil {
			o.controllerFilter = f
		}
	}
}

// MiddlewareFilter returns an Option that sets the controller filter.
//func MiddlewareFilter(f func(Middleware) bool) Option {
//	return func(o *options) {
//		if f != nil {
//			o.middlewareFilter = f
//		}
//	}
//}

// RequestTimeout returns an Option that sets the request timeout duration.
func RequestTimeout(d time.Duration) Option {
	return func(o *options) {
		o.requestTimeout = d
	}
}

// New creates a API server which has not started to accept requests yet.
func New(opt ...Option) *Server {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}

	s := &Server{
		opts:     opts,
		stopChan: make(chan struct{}),
	}

	if opts.mode == PROD {
		gin.SetMode(gin.ReleaseMode)
	}
	// gin.SetMode(gin.ReleaseMode) // disable gin log

	for _, c := range Controllers() {
		if !s.opts.controllerFilter(c) {
			s.controllerList = append(s.controllerList, c)
		}
	}

	//for _, m := range Middlewares() {
	//	if !s.opts.middlewareFilter(m) {
	//		s.middlewaresList = append(s.middlewaresList, m)
	//	}
	//}
	return s
}

// Startup bootstraps a server that contains controllers and middlewares.
func (s *Server) Startup(addr string) error {
	engine := gin.New()
	//if viper.GetBool("profile.enable") {
	//	pprof.Register(engine)
	//}
	// Setup global middleware
	// var authMiddleware, auditMiddleware Middleware
	//var authMiddleware Middleware
	//for _, m := range s.middlewaresList {
	//	// Different routers should use different AUTH.
	//	if m.Name() == "AUTH" {
	//		authMiddleware = m
	//		continue
	//	}
	//
	//	// if m.Name() == "AuditLog" {
	//	// 	auditMiddleware = m
	//	// 	continue
	//	// }
	//
	//	logger.Infof("Use middleware %s ", m.Name())
	//	engine.Use(gin.HandlerFunc(m.Handler()))
	//}

	v := s.opts.apiVersion
	if v == "" {
		v = defaultVersion
	}

	basePath := fmt.Sprintf("%s/%s", s.opts.basePath, v)
	rootGroup := engine.Group(basePath)

	noAuthHandler := func(c *gin.Context) {
		c.Set(utils.BAAS_API_AUTH_TYPE_KEY, utils.NoAuth)
	}

	basicAuthHandler := func(c *gin.Context) {
		c.Set(utils.BAAS_API_AUTH_TYPE_KEY, utils.BasicAuth)
	}

	tokenAuthHandler := func(c *gin.Context) {
		c.Set(utils.BAAS_API_AUTH_TYPE_KEY, utils.TokenAuth)
	}

	defaultRoles := viper.GetStringSlice("controller.routers.default.roles")
	if len(defaultRoles) == 0 {
		defaultRoles = utils.NotAuditRoles
	}

	for _, c := range s.controllerList {
		logger.Infof("Setup routes for controller %s", c.Name())
		groupPath := fmt.Sprintf("%s/%s", basePath, c.Name())
		rg := rootGroup.Group(c.Name())
		for _, r := range c.Routes() {
			logger.Infof(">> Setup API for %s %s%s", r.Method, groupPath, r.Path)
			//if authMiddleware == nil {
			//	rg.Handle(r.Method, r.Path, gin.HandlerFunc(r.Handler))
			//	continue
			//}

			var handlers []gin.HandlerFunc
			switch r.AuthType {
			case utils.NoAuth:
				handlers = append(handlers, noAuthHandler)
			case utils.BasicAuth:
				handlers = append(handlers, basicAuthHandler)
			default:
				handlers = append(handlers, tokenAuthHandler)
			}

			apiPath := groupPath
			if r.Path != "" {
				apiPath += r.Path
			}

			permissionInfo := &utils.APIPermissionInfo{
				APIPath:      apiPath,
				AllowedRoles: r.AllowedRoles,
			}

			if len(permissionInfo.AllowedRoles) == 0 {
				permissionInfo.AllowedRoles = defaultRoles
			}

			handlers = append(handlers, func(c *gin.Context) {
				c.Set(utils.BAAS_API_PERMISSION_KEY, permissionInfo)
			})

			// var auditInfo = r.AuditInfo
			// if auditInfo == nil {
			// 	auditInfo = &utils.AuditInfo{}
			// 	switch r.Method {
			// 	case "POST":
			// 		auditInfo.Required = true
			// 		auditInfo.Action = utils.AuditLogAction_Create
			// 	case "PUT", "PATCH":
			// 		auditInfo.Required = true
			// 		auditInfo.Action = utils.AuditLogAction_Update
			// 	case "DELETE":
			// 		auditInfo.Required = true
			// 		auditInfo.Action = utils.AuditLogAction_Delete
			// 	default:
			// 	}
			// }

			// handlers = append(handlers, func(c *gin.Context) {
			// 	c.Set(utils.BAAS_API_AUDIT_INFO_KEY, auditInfo)
			// })

			//if authMiddleware != nil {
			//	handlers = append(handlers, gin.HandlerFunc(authMiddleware.Handler()))
			//}

			// if auditMiddleware != nil {
			// 	handlers = append(handlers, gin.HandlerFunc(auditMiddleware.Handler()))
			// }

			handlers = append(handlers, gin.HandlerFunc(r.Handler))

			rg.Handle(
				r.Method,
				r.Path,
				handlers...,
			)
		}
	}

	//apiHTMLFile := "api/index.html"
	//cfgPath := os.Getenv("BAAS_CFG_PATH")
	//if cfgPath != "" {
	//	apiHTMLFile = filepath.Join(cfgPath, apiHTMLFile)
	//}
	//
	//engine.LoadHTMLFiles(apiHTMLFile)
	//engine.GET(basePath+"/index", func(c *gin.Context) {
	//	c.Header("content-type", "text/html;charset=utf-8")
	//	c.HTML(http.StatusOK, "index.html", nil)
	//})

	sv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  s.opts.requestTimeout,
		WriteTimeout: s.opts.requestTimeout,
	}

	s.server = sv

	logger.Infof("http Server s.server ReadTimeout:", s.server.ReadTimeout)
	logger.Infof("http Server s.server WriteTimeout:", s.server.WriteTimeout)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections.
func (s *Server) Shutdown(waitTime time.Duration) error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.TODO(), waitTime)
	defer cancel()
	return s.server.Shutdown(ctx)
}
