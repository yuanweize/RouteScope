package prober

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/yuanweize/RouteLens/pkg/logging"
	"golang.org/x/crypto/ssh"
)

// DefaultTestSize is the default amount of data to transfer (5MB)
const DefaultTestSize = 5 * 1024 * 1024

type SSHConfig struct {
	Host      string
	Port      int
	User      string
	Password  string
	KeyPath   string
	KeyText   string
	Timeout   time.Duration
	TestBytes int64 // How many bytes to test. If 0, uses DefaultTestSize
}

// SSHSpeedTester handles the SSH connection and speed measurement
type SSHSpeedTester struct {
	config SSHConfig
}

func NewSSHSpeedTester(cfg SSHConfig) *SSHSpeedTester {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.TestBytes == 0 {
		cfg.TestBytes = DefaultTestSize
	}
	return &SSHSpeedTester{config: cfg}
}

func (s *SSHSpeedTester) Run() (*SpeedResult, error) {
	target := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	logging.Info("ssh", "[SSH] Starting speed test for %s@%s", s.config.User, target)

	client, err := s.connect()
	if err != nil {
		// Categorize SSH connection errors for better diagnostics
		errMsg := err.Error()
		if strings.Contains(errMsg, "unable to authenticate") || strings.Contains(errMsg, "no supported methods") {
			logging.Error("ssh", "[SSH] Authentication failed for %s: invalid credentials or key", target)
		} else if strings.Contains(errMsg, "connection refused") {
			logging.Error("ssh", "[SSH] Connection refused by %s: port closed or firewall blocking", target)
		} else if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "i/o timeout") {
			logging.Error("ssh", "[SSH] Connection timeout for %s: host unreachable", target)
		} else if strings.Contains(errMsg, "no route to host") {
			logging.Error("ssh", "[SSH] No route to host %s: network unreachable", target)
		} else {
			logging.Error("ssh", "[SSH] Connection failed for %s: %v", target, err)
		}
		return nil, fmt.Errorf("ssh connection failed: %w", err)
	}
	defer client.Close()
	logging.Info("ssh", "[SSH] Connected to %s successfully", target)

	result := &SpeedResult{
		Timestamp: time.Now(),
	}

	// 1. Measure Download Speed (Remote -> Local)
	// Command: cat /dev/zero | head -c <TestBytes>
	logging.Debug("ssh", "[SSH] Starting download test for %s (%d bytes)", target, s.config.TestBytes)
	downSpeed, err := s.measureDownload(client)
	if err != nil {
		logging.Error("ssh", "[SSH] Download test failed for %s: %v", target, err)
		return nil, fmt.Errorf("download test failed: %w", err)
	}
	result.DownloadSpeed = downSpeed
	logging.Info("ssh", "[SSH] Download test for %s: %.2f Mbps", target, downSpeed)

	// 2. Measure Upload Speed (Local -> Remote)
	// Command: cat > /dev/null
	logging.Debug("ssh", "[SSH] Starting upload test for %s (%d bytes)", target, s.config.TestBytes)
	upSpeed, err := s.measureUpload(client)
	if err != nil {
		logging.Error("ssh", "[SSH] Upload test failed for %s: %v", target, err)
		return nil, fmt.Errorf("upload test failed: %w", err)
	}
	result.UploadSpeed = upSpeed
	logging.Info("ssh", "[SSH] Upload test for %s: %.2f Mbps", target, upSpeed)

	return result, nil
}

func (s *SSHSpeedTester) connect() (*ssh.Client, error) {
	auths := []ssh.AuthMethod{}
	if s.config.Password != "" {
		auths = append(auths, ssh.Password(s.config.Password))
	}
	if s.config.KeyText != "" {
		signer, err := ssh.ParsePrivateKey([]byte(s.config.KeyText))
		if err == nil {
			auths = append(auths, ssh.PublicKeys(signer))
		}
	}
	if s.config.KeyPath != "" {
		key, err := os.ReadFile(s.config.KeyPath)
		if err == nil {
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				auths = append(auths, ssh.PublicKeys(signer))
			}
		}
	}

	config := &ssh.ClientConfig{
		User:            s.config.User,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In production, we might want to store host keys
		Timeout:         s.config.Timeout,
	}

	target := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	return ssh.Dial("tcp", target, config)
}

func (s *SSHSpeedTester) measureDownload(client *ssh.Client) (float64, error) {
	session, err := client.NewSession()
	if err != nil {
		return 0, err
	}
	defer session.Close()

	cmd := fmt.Sprintf("cat /dev/zero | head -c %d", s.config.TestBytes)

	start := time.Now()
	reader, err := session.StdoutPipe()
	if err != nil {
		return 0, err
	}

	if err := session.Start(cmd); err != nil {
		return 0, err
	}

	n, err := io.Copy(io.Discard, reader)
	if err != nil && err != io.EOF {
		return 0, err
	}

	duration := time.Since(start)
	if err := session.Wait(); err != nil {
		// Verify exit code usually 0, but pipe might break early
	}

	// Mbps = (Bytes * 8) / (Seconds * 1000000)
	mbps := (float64(n) * 8) / (duration.Seconds() * 1000000)
	return mbps, nil
}

func (s *SSHSpeedTester) measureUpload(client *ssh.Client) (float64, error) {
	session, err := client.NewSession()
	if err != nil {
		return 0, err
	}
	defer session.Close()

	cmd := "cat > /dev/null"

	stdin, err := session.StdinPipe()
	if err != nil {
		return 0, err
	}

	if err := session.Start(cmd); err != nil {
		return 0, err
	}

	// Create a limit reader from /dev/zero but since we are writing, we need a generator
	// Ideally we could copy from local /dev/zero, but to be cross-platform we generate data
	start := time.Now()

	// Write chunks of zeroes
	chunkSize := 32 * 1024         // 32KB buffer
	buf := make([]byte, chunkSize) // Zeroed by default
	remaining := s.config.TestBytes

	var written int64
	for remaining > 0 {
		toWrite := int64(chunkSize)
		if remaining < toWrite {
			toWrite = remaining
		}

		n, err := stdin.Write(buf[:toWrite])
		written += int64(n)
		if err != nil {
			break
		}
		remaining -= int64(n)
	}
	stdin.Close() // Important to close stdin to signal EOF to remote cat

	duration := time.Since(start)
	session.Wait() // Wait for remote command to finish

	mbps := (float64(written) * 8) / (duration.Seconds() * 1000000)
	return mbps, nil
}
