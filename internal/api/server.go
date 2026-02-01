package api

import (
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuanweize/RouteLens/internal/auth"
	"github.com/yuanweize/RouteLens/internal/monitor"
	"github.com/yuanweize/RouteLens/pkg/storage"
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
	// CORS (Dev mode mostly, but kept for safety if run separately)
	s.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
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
	// TODO: Get real real-time status from Monitor Service cache
	// For now, return a mock
	c.JSON(http.StatusOK, gin.H{
		"targets": []gin.H{
			{
				"ip":           "8.8.8.8",
				"latency":      15.5,
				"loss":         0,
				"last_updated": time.Now(),
			},
		},
	})
}

func (s *Server) handleHistory(c *gin.Context) {
	s.handleStatus(c) // Use same mock for now
}

func (s *Server) handleProbe(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{"message": "Probe triggered"})
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
