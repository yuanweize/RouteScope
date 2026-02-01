package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/gin-gonic/gin"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/yuanweize/RouteLens/pkg/logging"
)

// Version information (set via ldflags)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

const (
	githubOwner = "yuanweize"
	githubRepo  = "RouteLens"
	githubSlug  = githubOwner + "/" + githubRepo
)

// getGitHubToken returns the GitHub token from environment if available
func getGitHubToken() string {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token
	}
	return ""
}

// UpdateCheckResponse represents the update check result
type UpdateCheckResponse struct {
	HasUpdate      bool   `json:"has_update"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version,omitempty"`
	ReleaseNotes   string `json:"release_notes,omitempty"`
	ReleaseURL     string `json:"release_url,omitempty"`
	PublishedAt    string `json:"published_at,omitempty"`
}

// SystemInfoResponse represents system information
type SystemInfoResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// handleSystemInfo returns current system information
func (s *Server) handleSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, SystemInfoResponse{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	})
}

// handleCheckUpdate checks for available updates on GitHub
func (s *Server) handleCheckUpdate(c *gin.Context) {
	logging.Info("update", "Checking for updates, current version: %s", Version)

	response := UpdateCheckResponse{
		CurrentVersion: Version,
		HasUpdate:      false,
	}

	client := &http.Client{Timeout: 15 * time.Second}
	token := getGitHubToken()

	// Strategy 1: Try GitHub Releases API (with token if available)
	latestVersion := s.tryGitHubReleasesAPI(client, token)

	// Strategy 2: Fallback to raw manifest file from master branch
	if latestVersion == "" {
		logging.Info("update", "Releases API failed, trying raw manifest fallback")
		latestVersion = s.tryRawManifestFallback(client)
	}

	// Strategy 3: Try selfupdate library (may also hit rate limit)
	if latestVersion == "" {
		logging.Info("update", "Manifest fallback failed, trying selfupdate library")
		latestVersion = s.trySelfupdateLibrary()
	}

	if latestVersion != "" {
		response.LatestVersion = latestVersion

		// Compare versions (handle both "v1.2.0" and "1.2.0" formats)
		if Version != "dev" {
			latestVer := strings.TrimPrefix(latestVersion, "v")
			currentVer := strings.TrimPrefix(Version, "v")
			latestSem, err1 := semver.Parse(latestVer)
			currentSem, err2 := semver.Parse(currentVer)
			if err1 == nil && err2 == nil {
				response.HasUpdate = latestSem.GT(currentSem)
			} else {
				logging.Warn("update", "Version parse error: latest=%v, current=%v", err1, err2)
			}
		}
		logging.Info("update", "Latest version: %s, current: %s, update available: %v",
			latestVersion, Version, response.HasUpdate)
	} else {
		logging.Warn("update", "All update check methods failed")
	}

	c.JSON(http.StatusOK, response)
}

// tryGitHubReleasesAPI attempts to get latest version from GitHub Releases API
func (s *Server) tryGitHubReleasesAPI(client *http.Client, token string) string {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubSlug)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "RouteLens-Updater")
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
		logging.Debug("update", "Using GitHub token for API request")
	}

	resp, err := client.Do(req)
	if err != nil {
		logging.Warn("update", "GitHub Releases API request failed: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		logging.Warn("update", "GitHub Releases API returned %d: %s", resp.StatusCode, string(body)[:min(200, len(body))])
		return ""
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if json.NewDecoder(resp.Body).Decode(&release) == nil && release.TagName != "" {
		logging.Info("update", "GitHub Releases API returned: %s", release.TagName)
		return release.TagName
	}
	return ""
}

// tryRawManifestFallback reads version from raw manifest file on master branch
func (s *Server) tryRawManifestFallback(client *http.Client) string {
	manifestURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/master/.github/.release-please-manifest.json", githubSlug)
	req, _ := http.NewRequest("GET", manifestURL, nil)
	req.Header.Set("User-Agent", "RouteLens-Updater")

	resp, err := client.Do(req)
	if err != nil {
		logging.Warn("update", "Raw manifest request failed: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logging.Warn("update", "Raw manifest returned %d", resp.StatusCode)
		return ""
	}

	var manifest map[string]string
	if json.NewDecoder(resp.Body).Decode(&manifest) == nil {
		if version, ok := manifest["."]; ok && version != "" {
			logging.Info("update", "Raw manifest returned version: %s", version)
			return "v" + strings.TrimPrefix(version, "v")
		}
	}
	return ""
}

// trySelfupdateLibrary uses the selfupdate library to detect latest version
func (s *Server) trySelfupdateLibrary() string {
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Filters: []string{
			fmt.Sprintf("routelens.*%s.*%s", runtime.GOOS, runtime.GOARCH),
		},
	})
	if err != nil {
		logging.Error("update", "Failed to create updater: %v", err)
		return ""
	}

	latest, found, err := updater.DetectLatest(githubSlug)
	if err != nil {
		logging.Warn("update", "Selfupdate library failed: %v", err)
		return ""
	}
	if found && latest != nil {
		return "v" + latest.Version.String()
	}
	return ""
}

// handlePerformUpdate downloads and applies the update
func (s *Server) handlePerformUpdate(c *gin.Context) {
	logging.Info("update", "Starting update process from version %s", Version)

	// Create updater
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Filters: []string{
			fmt.Sprintf("routelens_.*_%s_%s", runtime.GOOS, runtime.GOARCH),
		},
	})
	if err != nil {
		logging.Error("update", "Failed to create updater: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize updater"})
		return
	}

	// Get the current executable path
	exe, err := os.Executable()
	if err != nil {
		logging.Error("update", "Failed to get executable path: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get executable path"})
		return
	}

	// Check for latest release first
	latest, found, err := updater.DetectLatest(githubSlug)
	if err != nil {
		logging.Error("update", "Failed to detect latest version: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to detect latest version: %v", err)})
		return
	}

	if !found || latest == nil {
		logging.Warn("update", "No update available")
		c.JSON(http.StatusOK, gin.H{"message": "No update available", "updated": false})
		return
	}

	logging.Info("update", "Downloading version v%s...", latest.Version)

	// Perform the update
	err = updater.UpdateTo(latest, exe)
	if err != nil {
		logging.Error("update", "Update failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Update failed: %v", err)})
		return
	}

	logging.Info("update", "Update successful! New version: v%s. Restarting...", latest.Version)

	// Send success response before exiting
	c.JSON(http.StatusOK, gin.H{
		"message":     "Update successful. Service is restarting...",
		"updated":     true,
		"new_version": "v" + latest.Version.String(),
	})

	// Give time for response to be sent, then exit
	// Systemd will restart the service automatically
	go func() {
		time.Sleep(500 * time.Millisecond)
		logging.Info("update", "Exiting for restart...")
		os.Exit(0)
	}()
}
