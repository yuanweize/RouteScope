package api

import (
	"io/fs"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuanweize/RouteLens/internal/auth"
	"github.com/yuanweize/RouteLens/internal/monitor"
	"github.com/yuanweize/RouteLens/pkg/storage"
)

// Input validation patterns (Security: prevent command injection)
var (
	// Valid: IPv4, IPv6, domain names (RFC 1123)
	targetPattern = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z]{2,}$|^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$|^[a-fA-F0-9:]+$`)
)

type Server struct {
	router  *gin.Engine
	db      *storage.DB
	monitor *monitor.Service
	distFS  fs.FS
}

func NewServer(db *storage.DB, mon *monitor.Service, distFS fs.FS) *Server {
	r := gin.Default()
	s := &Server{
		router:  r,
		db:      db,
		monitor: mon,
		distFS:  distFS,
	}
	s.setupRoutes()
	return s
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) setupRoutes() {
	// CORS (Strict: same-origin only, unless explicitly configured)
	s.router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// Allow same-origin and localhost for dev
		if origin == "" || strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Setup enforcement: redirect all non-static requests to /setup until admin exists
	s.router.Use(func(c *gin.Context) {
		if s.db.HasAnyUser() {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/assets") || path == "/favicon.ico" || path == "/setup" || path == "/api/v1/setup" || path == "/api/v1/need-setup" {
			c.Next()
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, "/setup")
		c.Abort()
	})

	// Static Assets (Phase 8)
	if s.distFS != nil {
		dist, _ := fs.Sub(s.distFS, "dist")

		// Serve static files from /assets
		// Vite builds put all assets in /assets, so we map /assets to dist/assets
		assetsFS, _ := fs.Sub(dist, "assets")
		s.router.StaticFS("/assets", http.FS(assetsFS))

		// SPA Fallback: All other non-API routes serve index.html
		s.router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
				return
			}

			// Load index.html from embedded FS
			indexHTML, err := fs.ReadFile(dist, "index.html")
			if err != nil {
				c.String(http.StatusInternalServerError, "Error loading index.html")
				return
			}

			// Serve index.html with 200 OK for all SPA routes
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
		})
	}

	// Public API
	s.router.GET("/api/v1/need-setup", s.handleNeedSetup)
	s.router.POST("/api/v1/setup", s.handleSetup)
	s.router.POST("/login", s.handleLogin)

	// Protected API
	api := s.router.Group("/api/v1")
	api.Use(auth.AuthMiddleware())
	{
		api.GET("/status", s.handleStatus)
		api.GET("/history", s.handleHistory)
		api.GET("/trace", s.handleTrace)
		api.POST("/probe", s.handleProbe)
		api.POST("/user/password", s.handleUpdatePassword)

		// Target CRUD
		api.GET("/targets", s.handleGetTargets)
		api.POST("/targets", s.handleSaveTarget)
		api.DELETE("/targets/:id", s.handleDeleteTarget)
	}
}

// -- Handlers --

func (s *Server) handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	user, err := s.db.GetUser(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !auth.ComparePassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (s *Server) handleNeedSetup(c *gin.Context) {
	hasUser := s.db.HasAnyUser()
	c.JSON(http.StatusOK, gin.H{"need_setup": !hasUser})
}

func (s *Server) handleSetup(c *gin.Context) {
	if s.db.HasAnyUser() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Setup already completed"})
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, _ := auth.HashPassword(req.Password)
	user := &storage.User{
		Username: req.Username,
		Password: hashed,
	}
	if err := s.db.SaveUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setup successful"})
}

func (s *Server) handleUpdatePassword(c *gin.Context) {
	var req struct {
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For simplicity, update for "admin" or current user if we had identity in context
	// Currently all protected routes are for admin
	user, _ := s.db.GetUser("admin")
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		return
	}

	hashed, _ := auth.HashPassword(req.NewPassword)
	user.Password = hashed
	if err := s.db.SaveUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated"})
}

func (s *Server) handleStatus(c *gin.Context) {
	targets, err := s.db.GetTargets(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	status := make([]gin.H, 0, len(targets))
	for _, t := range targets {
		rec, recErr := s.db.GetLatestRecord(t.Address)
		if recErr != nil {
			status = append(status, gin.H{
				"target":     t,
				"latency":    0,
				"loss":       0,
				"speed_down": 0,
				"speed_up":   0,
				"updated_at": nil,
			})
			continue
		}
		status = append(status, gin.H{
			"target":     t,
			"latency":    rec.LatencyMs,
			"loss":       rec.PacketLoss,
			"speed_down": rec.SpeedDown,
			"speed_up":   rec.SpeedUp,
			"updated_at": rec.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"targets": status})
}

func (s *Server) handleHistory(c *gin.Context) {
	target := c.Query("target")
	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target is required"})
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	end := time.Now()
	start := end.Add(-6 * time.Hour)
	if startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = parsed
		}
	}
	if endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = parsed
		}
	}

	records, err := s.db.GetHistory(target, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

func (s *Server) handleProbe(c *gin.Context) {
	var req struct {
		Target string `json:"target"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Target == "" {
		req.Target = c.Query("target")
	}

	s.monitor.TriggerProbe(req.Target)
	c.JSON(http.StatusAccepted, gin.H{"message": "Probe triggered", "target": req.Target})
}

func (s *Server) handleTrace(c *gin.Context) {
	target := c.Query("target")
	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target is required"})
		return
	}

	rec, err := s.db.GetLatestTrace(target)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "trace not found"})
		return
	}

	c.Data(http.StatusOK, "application/json", rec.TraceJson)
}

func (s *Server) handleGetTargets(c *gin.Context) {
	targets, err := s.db.GetTargets(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, targets)
}

func (s *Server) handleSaveTarget(c *gin.Context) {
	var t storage.Target
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Security: Validate target address to prevent command injection
	if t.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address is required"})
		return
	}
	if !targetPattern.MatchString(t.Address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address format: only domain names and IP addresses allowed"})
		return
	}
	// Block shell metacharacters as extra safety
	if strings.ContainsAny(t.Address, ";|&$`\"'<>(){}[]") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address contains invalid characters"})
		return
	}

	if t.ProbeType == "" {
		t.ProbeType = storage.ProbeModeICMP
	}
	switch t.ProbeType {
	case storage.ProbeModeICMP, storage.ProbeModeHTTP, storage.ProbeModeSSH, storage.ProbeModeIPERF:
		// ok
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid probe_type"})
		return
	}
	if err := s.db.SaveTarget(&t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (s *Server) handleDeleteTarget(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := s.db.DeleteTarget(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Target deleted"})
}
