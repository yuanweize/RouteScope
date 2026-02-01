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
			if dbPath == "data/routescope.db" {
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
	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", "data/routescope.db", "Path to SQLite database")

	rootCmd.AddCommand(newServiceCmd())
	rootCmd.AddCommand(newUsersCmd())

	return rootCmd
}

func newUsersCmd() *cobra.Command {
	usersCmd := &cobra.Command{
		Use:   "users",
		Short: "User management",
	}

	resetPassCmd := &cobra.Command{
		Use:   "reset-password [username] [new_password]",
		Short: "Reset a user's password",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			username := args[0]
			password := args[1]

			db, err := storage.NewDB(dbPath)
			if err != nil {
				log.Fatalf("Failed to open DB: %v", err)
			}

			hashed, err := auth.HashPassword(password)
			if err != nil {
				log.Fatalf("Failed to hash password: %v", err)
			}

			user, err := db.GetUser(username)
			if err != nil {
				// Maybe try to create it?
				user = &storage.User{
					Username: username,
					Password: hashed,
				}
			} else {
				user.Password = hashed
			}

			if err := db.SaveUser(user); err != nil {
				log.Fatalf("Failed to save user: %v", err)
			}

			fmt.Printf("Password for user '%s' has been reset successfully.\n", username)
		},
	}

	usersCmd.AddCommand(resetPassCmd)
	return usersCmd
}

func newServiceCmd() *cobra.Command {
	serviceCmd := &cobra.Command{
		Use:   "service",
		Short: "Service management (install/uninstall)",
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install RouteLens as a systemd service (Linux only)",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS != "linux" {
				log.Fatal("Service installation is only supported on Linux")
			}

			installService()
		},
	}

	serviceCmd.AddCommand(installCmd)
	return serviceCmd
}

func installService() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

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

[Install]
WantedBy=multi-user.target
`, exeDir, exePath, port, dbPath, port, dbPath)

	servicePath := "/etc/systemd/system/routelens.service"
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

func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: failed to run %s %v: %v", name, args, err)
	}
}
