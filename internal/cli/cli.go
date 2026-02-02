package cli

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/yuanweize/RouteLens/internal/auth"
	"github.com/yuanweize/RouteLens/pkg/storage"
)

var (
	port   string
	dbPath string
)

func NewRootCmd(runServer func(port, dbPath string)) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "routelens",
		Short: "RouteLens - Modern Network Observation Platform",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Resolve default DB path if not provided
			if dbPath == "data/routelens.db" {
				// Check if env exists
				if env := os.Getenv("RS_DB_PATH"); env != "" {
					dbPath = env
				} else {
					// Default to local routelens.db if data/ dir doesn't exist
					if _, err := os.Stat("data"); os.IsNotExist(err) {
						dbPath = "routelens.db"
					}
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			runServer(port, dbPath)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "8080", "HTTP port to listen on")
	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", "data/routelens.db", "Path to SQLite database")

	rootCmd.AddCommand(newServiceCmd())
	rootCmd.AddCommand(newAdminCmd())

	return rootCmd
}

func newAdminCmd() *cobra.Command {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "Admin management",
	}

	resetPassCmd := &cobra.Command{
		Use:   "reset-password [new_password]",
		Short: "Reset the password for the system user",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			password := args[0]

			db, err := storage.NewDB(dbPath)
			if err != nil {
				log.Fatalf("Failed to open DB: %v", err)
			}

			hashed, err := auth.HashPassword(password)
			if err != nil {
				log.Fatalf("Failed to hash password: %v", err)
			}

			// Single-user system: get the first (and only) user
			user, err := db.GetFirstUser()
			if err != nil {
				log.Fatalf("No user found in database. Please complete initial setup first.")
			}

			// Use targeted update to avoid UNIQUE constraint issues
			if err := db.UpdateUserPassword(user.ID, hashed); err != nil {
				log.Fatalf("Failed to update password: %v", err)
			}

			fmt.Printf("Password for user '%s' has been reset successfully.\n", user.Username)
		},
	}

	adminCmd.AddCommand(resetPassCmd)
	return adminCmd
}

func newServiceCmd() *cobra.Command {
	serviceCmd := &cobra.Command{
		Use:   "service",
		Short: "Service management (install/uninstall)",
	}

	var force bool

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install RouteLens as a systemd service (Linux only)",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS != "linux" {
				log.Fatal("Service installation is only supported on Linux")
			}
			installService(force)
		},
	}
	installCmd.Flags().BoolVar(&force, "force", false, "Overwrite existing service file if present")

	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall RouteLens systemd service (Linux only)",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS != "linux" {
				log.Fatal("Service uninstallation is only supported on Linux")
			}
			uninstallService()
		},
	}

	serviceCmd.AddCommand(installCmd)
	serviceCmd.AddCommand(uninstallCmd)
	return serviceCmd
}

func installService(force bool) {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	geoPath := os.Getenv("RS_GEOIP_PATH")

	geoEnv := ""
	if geoPath != "" {
		geoEnv = fmt.Sprintf("Environment=RS_GEOIP_PATH=%s\n", geoPath)
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=RouteLens Monitoring Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=%s
ExecStart=%s --port %s --db %s
Restart=always
Environment=RS_HTTP_PORT=%s
Environment=RS_DB_PATH=%s
%s

[Install]
WantedBy=multi-user.target
`, exeDir, exePath, port, dbPath, port, dbPath, geoEnv)

	servicePath := "/etc/systemd/system/routelens.service"
	if !force {
		if _, err := os.Stat(servicePath); err == nil {
			log.Fatalf("Service file already exists at %s (use --force to overwrite)", servicePath)
		}
	}

	err = os.WriteFile(servicePath, []byte(serviceContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write service file: %v. Are you root?", err)
	}

	fmt.Println("Service file created at", servicePath)

	// Reload systemd
	runCmd("systemctl", "daemon-reload")
	runCmd("systemctl", "enable", "routelens")
	runCmd("systemctl", "start", "routelens")

	fmt.Println("RouteLens service installed and started successfully!")
}

func uninstallService() {
	servicePath := "/etc/systemd/system/routelens.service"
	runCmd("systemctl", "stop", "routelens")
	runCmd("systemctl", "disable", "routelens")

	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to remove service file: %v", err)
	}

	runCmd("systemctl", "daemon-reload")
	fmt.Println("RouteLens service uninstalled successfully!")
}

func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: failed to run %s %v: %v", name, args, err)
	}
}
