package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PlatformResult holds the outcome of a single platform posting attempt.
type PlatformResult struct {
	Platform    string
	ExecutionID string
	Success     bool
	Error       string
	PostedAt    time.Time
}

// DispatchToPlatforms fans out a video posting job to target platforms in parallel.
func DispatchToPlatforms(videoID int64, platforms []string, videoPath string, timeout time.Duration) []PlatformResult {
	if len(platforms) == 0 {
		return []PlatformResult{}
	}

	// Buffer equals platform count so no worker goroutine can block forever on send.
	results := make(chan PlatformResult, len(platforms))
	var wg sync.WaitGroup

	for _, platform := range platforms {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			defer func() {
				if recovered := recover(); recovered != nil {
					results <- PlatformResult{
						Platform: p,
						Success:  false,
						Error:    fmt.Sprintf("panic recovered: %v", recovered),
					}
				}
			}()

			results <- postToPlatform(p, videoID, videoPath, timeout)
		}(platform)
	}

	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("dispatcher closer panic recovered: %v", recovered)
			}
		}()
		wg.Wait()
		close(results)
	}()

	collected := make([]PlatformResult, 0, len(platforms))
	for result := range results {
		collected = append(collected, result)
	}
	return collected
}

func postToPlatform(platform string, videoID int64, videoPath string, timeout time.Duration) PlatformResult {
	// TODO: replace with real platform API client.
	log.Printf("STUB: would post video %d to %s from %s", videoID, platform, videoPath)

	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	select {
	case <-timer.C:
		return PlatformResult{
			Platform:    platform,
			ExecutionID: uuid.New().String(),
			Success:     true,
			PostedAt:    time.Now().UTC(),
		}
	case <-timeoutTimer.C:
		return PlatformResult{
			Platform: platform,
			Success:  false,
			Error:    "platform dispatch timeout",
			PostedAt: time.Now().UTC(),
		}
	}
}

// ResolvePlatforms expands "all" to every supported platform.
func ResolvePlatforms(videoPlatform string) []string {
	if videoPlatform == "all" {
		return []string{"instagram", "tiktok", "youtube"}
	}
	return []string{videoPlatform}
}
