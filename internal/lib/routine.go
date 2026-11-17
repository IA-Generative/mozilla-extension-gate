package lib

import "time"

func RunPeriodically(interval time.Duration, task func() error) {
	ticker := time.NewTicker(interval)
	_ = task()
	defer ticker.Stop()

	// Start a goroutine to periodically call the task function
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			_ = task()
		}
	}()
}
