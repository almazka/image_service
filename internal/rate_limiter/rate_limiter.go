package rate_limiter

import "sync"

type RateLimiter struct {
	mu                  sync.Mutex
	uploadDownloadCount int
	listFilesCount      int
	maxUploadDownload   int
	maxListFiles        int
}

func New(maxUploadDownload, maxListFiles int) *RateLimiter {
	return &RateLimiter{
		maxUploadDownload: maxUploadDownload,
		maxListFiles:      maxListFiles,
	}
}

func (rl *RateLimiter) CanUploadDownload() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.uploadDownloadCount < rl.maxUploadDownload {
		rl.uploadDownloadCount++
		return true
	}
	return false
}

func (rl *RateLimiter) ReleaseUploadDownload() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.uploadDownloadCount--
}

func (rl *RateLimiter) CanListFiles() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.listFilesCount < rl.maxListFiles {
		rl.listFilesCount++
		return true
	}
	return false
}

func (rl *RateLimiter) ReleaseListFiles() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.listFilesCount--
}
