package api

import (
	"encoding/json"
	"fmt"
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

	// Create updater with filter for correct binary
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Filters: []string{
			fmt.Sprintf("routelens.*%s.*%s", runtime.GOOS, runtime.GOARCH),
		},
	})
	if err != nil {
		logging.Error("update", "Failed to create updater: %v", err)
	}

	// Check for latest release via selfupdate library
	var latest *selfupdate.Release
	var found bool
	if updater != nil {
		latest, found, err = updater.DetectLatest(githubSlug)
		if err != nil {
			logging.Warn("update", "Failed to check for updates via updater: %v", err)
		}
	}

	// If no binary found, try to at least get the latest release version from GitHub API directly
	if !found || latest == nil {
		logging.Info("update", "No matching binary found, fetching latest release tag from GitHub API")
		
		// Direct GitHub API call to get latest release
		apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubSlug)
		client := &http.Client{Timeout: 10 * time.Second}
		req, _ := http.NewRequest("GET", apiURL, nil)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		
		resp, apiErr := client.Do(req)
		if apiErr == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			var release struct {
				TagName     string    `json:"tag_name"`
				Body        string    `json:"body"`
				HTMLURL     string    `json:"html_url"`
				PublishedAt time.Time `json:"published_at"`
			}
			if json.NewDecoder(resp.Body).Decode(&release) == nil {
				response.LatestVersion = release.TagName
				response.ReleaseNotes = release.Body
				response.ReleaseURL = release.HTMLURL
				if !release.PublishedAt.IsZero() {
					response.PublishedAt = release.PublishedAt.Format(time.RFC3339)
				}
				
				// Compare versions
				if Version != "dev" {
					latestVer := strings.TrimPrefix(release.TagName, "v")
					currentVer := strings.TrimPrefix(Version, "v")
					latestSem, err1 := semver.Parse(latestVer)
					currentSem, err2 := semver.Parse(currentVer)
					if err1 == nil && err2 == nil {
						response.HasUpdate = latestSem.GT(currentSem)
					}
				}
				logging.Info("update", "Latest release: %s, update available: %v", release.TagName, response.HasUpdate)
			}
		} else if apiErr != nil {
			logging.Warn("update", "GitHub API error: %v", apiErr)
		}
	} else if latest != nil {
		latestVer := latest.Version.String()
		response.LatestVersion = "v" + latestVer
		response.ReleaseNotes = latest.ReleaseNotes
		response.ReleaseURL = latest.URL
		if !latest.PublishedAt.IsZero() {
			response.PublishedAt = latest.PublishedAt.Format(time.RFC3339)
		}

		// Compare versions
		if Version == "dev" {
			response.HasUpdate = true
		} else {
			currentVer := strings.TrimPrefix(Version, "v")
			current, parseErr := semver.Parse(currentVer)
			if parseErr == nil {
				response.HasUpdate = latest.Version.GT(current)
			} else {
				logging.Warn("update", "Failed to parse current version '%s': %v", currentVer, parseErr)
				response.HasUpdate = true
			}
		}

		logging.Info("update", "Latest version: v%s, update available: %v", latestVer, response.HasUpdate)
	}

	c.JSON(http.StatusOK, response)
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
