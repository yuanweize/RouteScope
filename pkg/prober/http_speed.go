package prober

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPSpeedTester struct {
	URL string
}

func NewHTTPSpeedTester(url string) *HTTPSpeedTester {
	return &HTTPSpeedTester{URL: url}
}

func (h *HTTPSpeedTester) Run() (*SpeedResult, error) {
	start := time.Now()

	resp, err := http.Get(h.URL)
	if err != nil {
		return nil, fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http returned status: %s", resp.Status)
	}

	// Read body to measure speed
	// We use io.Discard to avoid memory overhead
	n, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	duration := time.Since(start)
	if duration == 0 {
		duration = time.Millisecond
	}

	// Calculate speed in Mbps
	// n is bytes, n*8 is bits, duration.Seconds() gives bps
	// bps / 1,000,000 = Mbps
	speedMbps := (float64(n) * 8) / (duration.Seconds() * 1000000)

	return &SpeedResult{
		DownloadSpeed: speedMbps,
		UploadSpeed:   0, // HTTP download doesn't measure upload speed easily
		Timestamp:     time.Now(),
	}, nil
}
