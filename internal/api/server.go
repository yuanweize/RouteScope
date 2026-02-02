package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oschwald/geoip2-golang"
	"github.com/yuanweize/RouteLens/internal/auth"
	"github.com/yuanweize/RouteLens/internal/monitor"
	"github.com/yuanweize/RouteLens/pkg/logging"
	"github.com/yuanweize/RouteLens/pkg/storage"
)

// Input validation patterns (Security: prevent command injection)
var (
	// Valid: IPv4, IPv6, domain names (RFC 1123)
	targetPattern = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z]{2,}$|^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$|^[a-fA-F0-9:]+$`)
)

// Rate limiter for login attempts: 5 attempts per IP per minute
var loginRateLimiter = NewRateLimiter(5, time.Minute)

// cleanSSHKeyInConfig cleans SSH key format in probe_config JSON
// Removes \r\n (Windows line endings) and normalizes to \n
func cleanSSHKeyInConfig(configJSON string) string {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return configJSON // Return as-is if not valid JSON
	}

	// Clean ssh_key field if present
	if key, ok := config["ssh_key"].(string); ok && key != "" {
		// Replace \r\n with \n, then remove any standalone \r
		key = strings.ReplaceAll(key, "\r\n", "\n")
		key = strings.ReplaceAll(key, "\r", "")
		// Trim whitespace
		key = strings.TrimSpace(key)
		// Ensure key ends with newline (required by SSH)
		if !strings.HasSuffix(key, "\n") {
			key += "\n"
		}
		config["ssh_key"] = key
	}

	result, err := json.Marshal(config)
	if err != nil {
		return configJSON
	}
	return string(result)
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

type Server struct {
	router   *gin.Engine
	db       *storage.DB
	monitor  *monitor.Service
	distFS   fs.FS
	dbPath   string
	settings SystemSettings
}

func NewServer(db *storage.DB, mon *monitor.Service, distFS fs.FS, dbPath string) *Server {
	r := gin.Default()
	s := &Server{
		router:  r,
		db:      db,
		monitor: mon,
		distFS:  distFS,
		dbPath:  dbPath,
		settings: SystemSettings{
			RetentionDays:     30,
			SpeedTestInterval: 5,
			PingInterval:      30,
		},
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

		// Serve static files from /assets with long cache (files have content hash)
		// Vite builds put all assets in /assets with hashed filenames like index-CugdYbR7.js
		assetsFS, _ := fs.Sub(dist, "assets")
		s.router.Group("/assets").Use(func(c *gin.Context) {
			// Hashed assets: cache for 1 year (immutable)
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			c.Next()
		}).StaticFS("/", http.FS(assetsFS))

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

			// index.html: no-cache to ensure browser always checks for updates
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
		})
	}

	// Public API
	s.router.GET("/api/v1/need-setup", s.handleNeedSetup)
	s.router.POST("/api/v1/setup", s.handleSetup)
	// Login with rate limiting: 5 attempts per IP per minute
	s.router.POST("/login", LoginRateLimitMiddleware(loginRateLimiter), s.handleLogin)
	s.router.GET("/api/v1/system/info", s.handleSystemInfo)      // Public: version info is not sensitive
	s.router.GET("/api/v1/system/releases", s.handleGetReleases) // Public: GitHub releases info

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

		// System Logs
		api.GET("/logs", s.handleGetLogs)

		// System Update (Self-Update) - Protected
		api.GET("/system/check-update", s.handleCheckUpdate)
		api.POST("/system/update", s.handlePerformUpdate)

		// Database Management - Protected
		api.GET("/system/database/stats", s.handleGetDatabaseStats)
		api.POST("/system/database/clean", s.handleCleanDatabase)
		api.POST("/system/database/vacuum", s.handleVacuumDatabase)
		api.GET("/system/settings", s.handleGetSettings)
		api.POST("/system/settings", s.handleSaveSettings)

		// GeoIP Management - Protected
		api.GET("/system/geoip/status", s.handleGetGeoIPStatus)
		api.POST("/system/geoip/update", s.handleUpdateGeoIP)
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

	token, err := auth.GenerateToken(user.Username)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate username
	if len(req.Username) < 3 || len(req.Username) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be 3-32 characters"})
		return
	}
	// Validate password
	if len(req.Password) < 6 || len(req.Password) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be 6-72 characters"})
		return
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		logging.Error("setup", "Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}
	user := &storage.User{
		Username: req.Username,
		Password: hashed,
	}
	if err := s.db.SaveUser(user); err != nil {
		logging.Error("setup", "Failed to save user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setup successful"})
}

func (s *Server) handleUpdatePassword(c *gin.Context) {
	var req struct {
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate password
	if len(req.NewPassword) < 6 || len(req.NewPassword) > 72 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be 6-72 characters"})
		return
	}

	// Get the first (and only) user in the system
	// This is a single-user system, so we update the only existing user
	user, err := s.db.GetFirstUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No user found in system"})
		return
	}

	// Hash new password and update the existing record
	hashed, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		logging.Error("password", "Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}
	if err := s.db.UpdateUserPassword(user.ID, hashed); err != nil {
		logging.Error("password", "Failed to update password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated"})
}

func (s *Server) handleStatus(c *gin.Context) {
	targets, err := s.db.GetTargets(false)
	if err != nil {
		logging.Error("api", "Failed to get targets: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch targets"})
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
		logging.Error("api", "Failed to get history for %s: %v", target, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
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

	// Validate target if provided
	if req.Target != "" && !targetPattern.MatchString(req.Target) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target format"})
		return
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

	// Check if language localization is needed
	lang := c.Query("lang")
	if lang == "" || strings.HasPrefix(lang, "zh") {
		// Default: return raw JSON (Chinese)
		c.Data(http.StatusOK, "application/json", rec.TraceJson)
		return
	}

	// For English, swap city/subdiv/country with their _en versions
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.TraceJson, &payload); err != nil {
		c.Data(http.StatusOK, "application/json", rec.TraceJson)
		return
	}

	if hops, ok := payload["hops"].([]interface{}); ok {
		for _, hopRaw := range hops {
			if hop, ok := hopRaw.(map[string]interface{}); ok {
				// Use English fields if available
				if cityEN, ok := hop["city_en"].(string); ok && cityEN != "" {
					hop["city"] = cityEN
				}
				if subdivEN, ok := hop["subdiv_en"].(string); ok && subdivEN != "" {
					hop["subdiv"] = subdivEN
				}
				if countryEN, ok := hop["country_en"].(string); ok && countryEN != "" {
					hop["country"] = countryEN
				}
			}
		}
	}

	localizedJson, err := json.Marshal(payload)
	if err != nil {
		c.Data(http.StatusOK, "application/json", rec.TraceJson)
		return
	}
	c.Data(http.StatusOK, "application/json", localizedJson)
}

func (s *Server) handleGetTargets(c *gin.Context) {
	targets, err := s.db.GetTargets(false)
	if err != nil {
		logging.Error("api", "Failed to get targets: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch targets"})
		return
	}
	c.JSON(http.StatusOK, targets)
}

func (s *Server) handleSaveTarget(c *gin.Context) {
	var t storage.Target
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
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

	// Clean SSH key in probe_config: remove \r\n and normalize to \n
	if t.ProbeConfig != "" && t.ProbeType == storage.ProbeModeSSH {
		t.ProbeConfig = cleanSSHKeyInConfig(t.ProbeConfig)
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

	// Distinguish between Create (ID=0) and Update (ID>0)
	if t.ID == 0 {
		// Create new target
		if err := s.db.CreateTarget(&t); err != nil {
			// Handle duplicate address error gracefully
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				c.JSON(http.StatusConflict, gin.H{"error": "Target with this address already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Update existing target - verify it exists first
		existing, err := s.db.GetTargetByID(t.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Target not found"})
			return
		}
		// Preserve created_at from existing record
		t.CreatedAt = existing.CreatedAt
		if err := s.db.UpdateTarget(&t); err != nil {
			logging.Error("api", "Failed to update target %d: %v", t.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update target"})
			return
		}
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
		logging.Error("api", "Failed to delete target %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete target"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Target deleted"})
}

func (s *Server) handleGetLogs(c *gin.Context) {
	// Parse query parameters
	linesStr := c.DefaultQuery("lines", "100")
	lines, _ := strconv.Atoi(linesStr)
	if lines <= 0 || lines > 1000 {
		lines = 100
	}

	levelFilter := c.Query("level") // e.g., "ERROR", "WARN,ERROR"

	logger := logging.GetGlobalLogger()
	var entries []logging.LogEntry

	if levelFilter != "" {
		// Parse comma-separated levels
		levelStrs := strings.Split(levelFilter, ",")
		var levels []logging.LogLevel
		for _, ls := range levelStrs {
			ls = strings.TrimSpace(strings.ToUpper(ls))
			switch ls {
			case "DEBUG":
				levels = append(levels, logging.LevelDebug)
			case "INFO":
				levels = append(levels, logging.LevelInfo)
			case "WARN":
				levels = append(levels, logging.LevelWarn)
			case "ERROR":
				levels = append(levels, logging.LevelError)
			}
		}
		entries = logger.GetByLevel(levels...)
	} else {
		entries = logger.GetLast(lines)
	}

	// Return only the last N entries after filtering
	if len(entries) > lines {
		entries = entries[len(entries)-lines:]
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  entries,
		"count": len(entries),
	})
}

// --- Database Management Handlers ---

func (s *Server) handleGetDatabaseStats(c *gin.Context) {
	stats, err := s.db.GetDatabaseStats(s.dbPath, s.settings.RetentionDays)
	if err != nil {
		logging.Error("api", "Failed to get database stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database statistics"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (s *Server) handleCleanDatabase(c *gin.Context) {
	var req struct {
		Days int `json:"days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Days = s.settings.RetentionDays
	}
	if req.Days < 1 {
		req.Days = 7
	}

	deleted, err := s.db.CleanOldRecords(req.Days)
	if err != nil {
		logging.Error("database", "Failed to clean old records: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clean database"})
		return
	}

	logging.Info("database", "Cleaned %d old records (older than %d days)", deleted, req.Days)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Deleted %d records older than %d days", deleted, req.Days),
		"deleted": deleted,
	})
}

func (s *Server) handleVacuumDatabase(c *gin.Context) {
	if err := s.db.VacuumDatabase(); err != nil {
		logging.Error("database", "Failed to vacuum database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to vacuum database"})
		return
	}

	logging.Info("database", "Database vacuumed successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Database vacuumed successfully"})
}

// Settings management
type SystemSettings struct {
	RetentionDays     int `json:"retention_days"`
	SpeedTestInterval int `json:"speed_test_interval_minutes"`
	PingInterval      int `json:"ping_interval_seconds"`
}

func (s *Server) handleGetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, s.settings)
}

func (s *Server) handleSaveSettings(c *gin.Context) {
	var req SystemSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate
	if req.RetentionDays < 1 {
		req.RetentionDays = 7
	}
	if req.RetentionDays > 365 {
		req.RetentionDays = 365
	}
	if req.SpeedTestInterval < 1 {
		req.SpeedTestInterval = 5
	}
	if req.PingInterval < 10 {
		req.PingInterval = 30
	}

	s.settings = req
	// TODO: Persist settings to database or config file
	logging.Info("settings", "Settings updated: retention=%d days, speed=%d min, ping=%d sec",
		req.RetentionDays, req.SpeedTestInterval, req.PingInterval)
	c.JSON(http.StatusOK, s.settings)
}

// GeoIP Status Response
type GeoIPStatus struct {
	Available     bool   `json:"available"`
	Path          string `json:"path"`
	SizeBytes     int64  `json:"size_bytes"`
	SizeHuman     string `json:"size_human"`
	ModTime       string `json:"mod_time"`
	LastUpdated   string `json:"last_updated"`
	DatabaseType  string `json:"database_type"`
	BuildEpoch    int64  `json:"build_epoch"`
	BuildTime     string `json:"build_time"`
	IPVersion     int    `json:"ip_version"`
	NodeCount     uint   `json:"node_count"`
	RecordSize    uint   `json:"record_size"`
	BinaryVersion string `json:"binary_version"`
	Description   string `json:"description"`
}

func (s *Server) handleGetGeoIPStatus(c *gin.Context) {
	geoipPath := filepath.Join(filepath.Dir(s.dbPath), "geoip", "GeoLite2-City.mmdb")

	status := GeoIPStatus{
		Available: false,
		Path:      geoipPath,
	}

	if info, err := os.Stat(geoipPath); err == nil {
		status.Available = true
		status.SizeBytes = info.Size()
		status.SizeHuman = formatBytes(info.Size())
		status.ModTime = info.ModTime().Format(time.RFC3339)
		status.LastUpdated = info.ModTime().Format("2006-01-02 15:04:05")

		// Read MMDB metadata
		if db, err := geoip2.Open(geoipPath); err == nil {
			defer db.Close()
			meta := db.Metadata()
			status.DatabaseType = meta.DatabaseType
			status.BuildEpoch = int64(meta.BuildEpoch)
			status.BuildTime = time.Unix(int64(meta.BuildEpoch), 0).Format("2006-01-02")
			status.IPVersion = int(meta.IPVersion)
			status.NodeCount = meta.NodeCount
			status.RecordSize = meta.RecordSize
			status.BinaryVersion = fmt.Sprintf("%d.%d", meta.BinaryFormatMajorVersion, meta.BinaryFormatMinorVersion)
			if desc, ok := meta.Description["en"]; ok {
				status.Description = desc
			}
		}
	}

	c.JSON(http.StatusOK, status)
}

func (s *Server) handleUpdateGeoIP(c *gin.Context) {
	geoipDir := filepath.Join(filepath.Dir(s.dbPath), "geoip")
	geoipPath := filepath.Join(geoipDir, "GeoLite2-City.mmdb")

	// Create directory if not exists
	if err := os.MkdirAll(geoipDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create geoip directory"})
		return
	}

	// Download from a public mirror (using db-ip.com free database as fallback)
	// MaxMind requires license key, so we use db-ip.com's free City database
	downloadURL := "https://github.com/P3TERX/GeoLite.mmdb/releases/latest/download/GeoLite2-City.mmdb"

	logging.Info("geoip", "Downloading GeoIP database from %s", downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Download failed: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Download failed: HTTP %d", resp.StatusCode)})
		return
	}

	// Write to temp file first
	tmpPath := geoipPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create temp file: %v", err)})
		return
	}

	written, err := io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(tmpPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to write file: %v", err)})
		return
	}

	// Replace old file
	if err := os.Rename(tmpPath, geoipPath); err != nil {
		os.Remove(tmpPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to replace file: %v", err)})
		return
	}

	logging.Info("geoip", "GeoIP database updated successfully: %s (%s)", geoipPath, formatBytes(written))

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "GeoIP database updated successfully",
		"size_bytes": written,
		"size_human": formatBytes(written),
	})
}
