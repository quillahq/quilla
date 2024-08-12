package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/negroni"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	scas "github.com/qiangmzsx/string-adapter/v2"
	"github.com/quilla-hq/quilla/approvals"
	"github.com/quilla-hq/quilla/constants"
	"github.com/quilla-hq/quilla/internal/k8s"
	"github.com/quilla-hq/quilla/pkg/auth"
	"github.com/quilla-hq/quilla/pkg/store"
	"github.com/quilla-hq/quilla/provider"
	"github.com/quilla-hq/quilla/provider/kubernetes"
	"github.com/quilla-hq/quilla/types"
	"github.com/quilla-hq/quilla/version"

	log "github.com/sirupsen/logrus"
)

// Opts - http server options
type Opts struct {
	Port int

	// available providers
	Providers provider.Providers

	ApprovalManager approvals.Manager

	Authenticator auth.Authenticator

	GRC *k8s.GenericResourceCache

	KubernetesClient kubernetes.Implementer

	Store store.Store

	UIDir string

	AuthenticatedWebhooks bool

	RBACEnabled bool
}

// TriggerServer - webhook trigger & healthcheck server
type TriggerServer struct {
	grc              *k8s.GenericResourceCache
	kubernetesClient kubernetes.Implementer

	providers        provider.Providers
	approvalsManager approvals.Manager
	port             int
	server           *http.Server
	router           *mux.Router

	store         store.Store
	authenticator auth.Authenticator

	uiDir string

	authenticatedWebhooks bool

	e *casbin.Enforcer
}

// NewTriggerServer - create new HTTP trigger based server
func NewTriggerServer(opts *Opts) *TriggerServer {
	var e *casbin.Enforcer
	if opts.RBACEnabled {
		m, err := model.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && regexMatch(r.act, p.act)
`)
		if err != nil {
			panic(err)
		}

		sa := scas.NewAdapter(os.Getenv(constants.EnvRBACPolicy))

		e, err = casbin.NewEnforcer(m, sa)
		if err != nil {
			panic(err)
		}
	}
	return &TriggerServer{
		port:                  opts.Port,
		grc:                   opts.GRC,
		kubernetesClient:      opts.KubernetesClient,
		providers:             opts.Providers,
		approvalsManager:      opts.ApprovalManager,
		router:                mux.NewRouter(),
		authenticator:         opts.Authenticator,
		store:                 opts.Store,
		uiDir:                 opts.UIDir,
		authenticatedWebhooks: opts.AuthenticatedWebhooks,
		e:                     e,
	}
}

// Start - start server
func (s *TriggerServer) Start() error {

	s.registerRoutes(s.router)

	n := negroni.New(negroni.NewRecovery())
	n.Use(negroni.HandlerFunc(corsHeadersMiddleware))
	n.UseHandler(s.router)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: n,
	}

	log.WithFields(log.Fields{
		"port": s.port,
	}).Info("webhook trigger server starting...")

	return s.server.ListenAndServe()
}

// Stop - stop webhook server
func (s *TriggerServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.server.Shutdown(ctx)
}

func getID(req *http.Request) string {
	return mux.Vars(req)["id"]
}

func (s *TriggerServer) registerRoutes(mux *mux.Router) {

	if os.Getenv("DEBUG") == "true" {
		DebugHandler{}.AddRoutes(mux)
	}

	s.registerWebhookRoutes(mux)

	// health endpoint for k8s to be happy
	mux.HandleFunc("/healthz", s.healthHandler).Methods("GET", "OPTIONS")
	// version handler
	mux.HandleFunc("/version", s.versionHandler).Methods("GET", "OPTIONS")

	mux.Handle("/metrics", promhttp.Handler())

	if s.authenticator.Enabled() {
		log.Info("authentication enabled, setting up admin HTTP handlers")
		// auth
		mux.HandleFunc("/v1/auth/login", s.loginHandler).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/auth/info", s.requireAdminAuthorization(s.userInfoHandler)).Methods("GET", "OPTIONS")
		mux.HandleFunc("/v1/auth/user", s.requireAdminAuthorization(s.userInfoHandler)).Methods("GET", "OPTIONS")
		mux.HandleFunc("/v1/auth/logout", s.requireAdminAuthorization(s.logoutHandler)).Methods("POST", "GET", "OPTIONS")
		mux.HandleFunc("/v1/auth/refresh", s.requireAdminAuthorization(s.refreshHandler)).Methods("GET", "OPTIONS")

		// approvals
		mux.HandleFunc("/v1/approvals", s.requireAdminAuthorization(s.requireRBAC(s.approvalsHandler, "approvals", "read"))).Methods("GET", "OPTIONS")
		// approving/rejecting
		mux.HandleFunc("/v1/approvals", s.requireAdminAuthorization(s.requireRBAC(s.approvalApproveHandler, "approvals", "write"))).Methods("POST", "OPTIONS")
		// updating required approvals count
		mux.HandleFunc("/v1/approvals", s.requireAdminAuthorization(s.requireRBAC(s.approvalSetHandler, "approvals", "write"))).Methods("PUT", "OPTIONS")

		// available resources
		mux.HandleFunc("/v1/resources", s.requireAdminAuthorization(s.requireRBAC(s.resourcesHandler, "resources", "read"))).Methods("GET", "OPTIONS")

		mux.HandleFunc("/v1/policies", s.requireAdminAuthorization(s.requireRBAC(s.policyUpdateHandler, "policies", "write"))).Methods("PUT", "OPTIONS")

		// tracked images
		mux.HandleFunc("/v1/tracked", s.requireAdminAuthorization(s.requireRBAC(s.trackedHandler, "tracked", "read"))).Methods("GET", "OPTIONS")
		mux.HandleFunc("/v1/tracked", s.requireAdminAuthorization(s.requireRBAC(s.trackSetHandler, "tracked", "write"))).Methods("PUT", "OPTIONS")

		// status
		mux.HandleFunc("/v1/audit", s.requireAdminAuthorization(s.requireRBAC(s.adminAuditLogHandler, "audit", "read"))).Methods("GET", "OPTIONS")
		mux.HandleFunc("/v1/stats", s.requireAdminAuthorization(s.requireRBAC(s.statsHandler, "stats", "read"))).Methods("GET", "OPTIONS")

		// config
		mux.HandleFunc("/v1/config", s.configHandler).Methods("GET")
		mux.HandleFunc("/v1/me", s.requireAdminAuthorization(s.userHandler)).Methods("GET")

		if s.uiDir != "" {
			spa := SpaHandler{StaticPath: s.uiDir, IndexPath: "index.html"}
			mux.PathPrefix("/").Handler(spa)
		}
	} else {
		log.Info("authentication is not enabled, admin HTTP handlers are not initialized")
	}

}

type SpaHandler struct {
	StaticPath string
	IndexPath  string
}

func (h SpaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(h.StaticPath, r.URL.Path)

	fi, err := os.Stat(path)
	if os.IsNotExist(err) || fi.IsDir() {
		http.ServeFile(w, r, filepath.Join(h.StaticPath, h.IndexPath))
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.FileServer(http.Dir(h.StaticPath)).ServeHTTP(w, r)
}

func (s *TriggerServer) registerWebhookRoutes(mux *mux.Router) {

	if s.authenticatedWebhooks {
		mux.HandleFunc("/v1/webhooks/native", s.requireAdminAuthorization(s.nativeHandler)).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/dockerhub", s.requireAdminAuthorization(s.dockerHubHandler)).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/jfrog", s.requireAdminAuthorization(s.jfrogHandler)).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/quay", s.requireAdminAuthorization(s.quayHandler)).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/azure", s.requireAdminAuthorization(s.azureHandler)).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/github", s.requireAdminAuthorization(s.githubHandler)).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/harbor", s.requireAdminAuthorization(s.harborHandler)).Methods("POST", "OPTIONS")

		// Docker registry notifications, used by Docker, Gitlab, Harbor
		// https://docs.docker.com/registry/notifications/
		//https://docs.gitlab.com/ee/administration/container_registry.html#configure-container-registry-notifications
		mux.HandleFunc("/v1/webhooks/registry", s.registryNotificationHandler).Methods("POST", "OPTIONS")
	} else {
		mux.HandleFunc("/v1/webhooks/native", s.nativeHandler).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/dockerhub", s.dockerHubHandler).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/jfrog", s.jfrogHandler).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/quay", s.quayHandler).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/azure", s.azureHandler).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/github", s.githubHandler).Methods("POST", "OPTIONS")
		mux.HandleFunc("/v1/webhooks/harbor", s.harborHandler).Methods("POST", "OPTIONS")

		// Docker registry notifications, used by Docker, Gitlab, Harbor
		// https://docs.docker.com/registry/notifications/
		//https://docs.gitlab.com/ee/administration/container_registry.html#configure-container-registry-notifications
		mux.HandleFunc("/v1/webhooks/registry", s.registryNotificationHandler).Methods("POST", "OPTIONS")
	}
}

func (s *TriggerServer) healthHandler(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusOK)
}

func (s *TriggerServer) versionHandler(resp http.ResponseWriter, req *http.Request) {
	v := version.GetquillaVersion()

	encoded, err := json.Marshal(v)
	if err != nil {
		log.WithError(err).Error("trigger.http: failed to marshal version")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Write(encoded)
}

func (s *TriggerServer) trigger(event types.Event) error {
	return s.providers.Submit(event)
}

func response(obj interface{}, statusCode int, err error, resp http.ResponseWriter, req *http.Request) {
	// Check for an error

	if err != nil {

		code := 500
		errMsg := err.Error()
		if strings.Contains(errMsg, "Permission denied") {
			code = 403
		}
		resp.WriteHeader(code)
		resp.Write([]byte(err.Error()))
		return
	}

	// Write out the JSON object
	if obj != nil {

		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(statusCode)

		// Set up the pipe to write data directly into the Reader.
		pr, pw := io.Pipe()

		// Write JSON-encoded data to the Writer end of the pipe.
		// Write in a separate concurrent goroutine, and remember
		// to Close the PipeWriter, to signal to the paired PipeReader
		// that we’re done writing.
		go func() {
			pw.CloseWithError(json.NewEncoder(pw).Encode(obj))
		}()

		io.Copy(resp, pr)

		// encoding/json library has a specific bug(feature) to turn empty slices into json null object,
		// let's make an empty array instead
		// resp.Write(buf)
	}
}

// corsHeadersMiddleware - cors middleware
func corsHeadersMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	rw.Header().Set("Access-Control-Expose-Headers", "Authorization")
	rw.Header().Set("Access-Control-Request-Headers", "Authorization")

	if r.Method == "OPTIONS" {
		rw.WriteHeader(200)
		return
	}

	next(rw, r)
}

type UserInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Username      string `json:"username"`
	Avatar        string `json:"avatar"`
	Status        int    `json:"status"`
	LastLoginIP   string `json:"last_login_ip"`
	LastLoginTime int64  `json:"last_login_time"`
	RoleID        string `json:"role_id"`
}

func (s *TriggerServer) userInfoHandler(resp http.ResponseWriter, req *http.Request) {

	user := auth.GetAccountFromCtx(req.Context())

	ui := UserInfo{
		ID:            "1",
		Name:          user.Username,
		Avatar:        "",
		Status:        1,
		LastLoginIP:   "",
		LastLoginTime: time.Now().Unix(),
		RoleID:        "admin",
	}

	response(&ui, 200, nil, resp, req)
}

type APIResponse struct {
	Status string `json:"status"`
}

func indexHandler(uiDir string) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, uiDir+"/index.html")
	}

	return http.HandlerFunc(fn)
}
