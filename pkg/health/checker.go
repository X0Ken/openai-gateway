package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Status represents the health status of a channel
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// ChannelHealth represents the health state of a channel
type ChannelHealth struct {
	ChannelID           int64     `json:"channel_id"`
	Status              Status    `json:"status"`
	LastChecked         time.Time `json:"last_checked"`
	LastError           string    `json:"last_error,omitempty"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
}

// Checker manages health checks for channels
type Checker struct {
	mu       sync.RWMutex
	statuses map[int64]*ChannelHealth
	interval time.Duration
	timeout  time.Duration
	stopCh   chan struct{}
}

// NewChecker creates a new health checker
func NewChecker(interval, timeout time.Duration) *Checker {
	return &Checker{
		statuses: make(map[int64]*ChannelHealth),
		interval: interval,
		timeout:  timeout,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the health check loop
func (c *Checker) Start() {
	go c.loop()
}

// Stop stops the health check loop
func (c *Checker) Stop() {
	close(c.stopCh)
}

// loop runs the health check loop
func (c *Checker) loop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.checkAll()
		case <-c.stopCh:
			return
		}
	}
}

// checkAll performs health checks on all registered channels
func (c *Checker) checkAll() {
	c.mu.RLock()
	channelIDs := make([]int64, 0, len(c.statuses))
	for id := range c.statuses {
		channelIDs = append(channelIDs, id)
	}
	c.mu.RUnlock()

	for _, id := range channelIDs {
		c.checkChannel(id)
	}
}

// checkChannel performs a health check on a single channel
func (c *Checker) checkChannel(channelID int64) {
	// This would normally check the actual channel endpoint
	// For now, we'll just update the timestamp
	c.mu.Lock()
	defer c.mu.Unlock()

	if status, exists := c.statuses[channelID]; exists {
		status.LastChecked = time.Now()
	}
}

// RegisterChannel registers a channel for health checking
func (c *Checker) RegisterChannel(channelID int64, baseURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.statuses[channelID] = &ChannelHealth{
		ChannelID:   channelID,
		Status:      StatusUnknown,
		LastChecked: time.Now(),
	}
}

// UnregisterChannel removes a channel from health checking
func (c *Checker) UnregisterChannel(channelID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.statuses, channelID)
}

// GetStatus returns the health status of a channel
func (c *Checker) GetStatus(channelID int64) *ChannelHealth {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.statuses[channelID]
}

// GetAllStatuses returns the health status of all channels
func (c *Checker) GetAllStatuses() []*ChannelHealth {
	c.mu.RLock()
	defer c.mu.RUnlock()

	statuses := make([]*ChannelHealth, 0, len(c.statuses))
	for _, status := range c.statuses {
		statuses = append(statuses, status)
	}

	return statuses
}

// UpdateStatus updates the health status of a channel (passive detection)
func (c *Checker) UpdateStatus(channelID int64, healthy bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	status, exists := c.statuses[channelID]
	if !exists {
		return
	}

	status.LastChecked = time.Now()

	if healthy {
		status.Status = StatusHealthy
		status.ConsecutiveFailures = 0
		status.LastError = ""
	} else {
		status.ConsecutiveFailures++
		if status.ConsecutiveFailures >= 3 {
			status.Status = StatusUnhealthy
		}
		if err != nil {
			status.LastError = err.Error()
		}
	}
}

// CheckEndpoint performs an HTTP health check on an endpoint
func CheckEndpoint(ctx context.Context, url string, timeout time.Duration) error {
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}
