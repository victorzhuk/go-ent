package execution

import (
	"fmt"
	"time"
)

// ResourceLimits defines execution resource constraints.
type ResourceLimits struct {
	// MaxMemoryMB is the maximum memory in megabytes.
	MaxMemoryMB int

	// MaxCPUTime is the maximum CPU time allowed.
	MaxCPUTime time.Duration

	// MaxExecTime is the maximum wall-clock execution time.
	MaxExecTime time.Duration
}

// DefaultResourceLimits returns safe default limits for sandbox execution.
func DefaultResourceLimits() ResourceLimits {
	return ResourceLimits{
		MaxMemoryMB: 128,              // 128MB memory limit
		MaxCPUTime:  30 * time.Second, // 30s CPU time
		MaxExecTime: 60 * time.Second, // 60s wall-clock time
	}
}

// Sandbox isolates untrusted code execution with resource limits.
type Sandbox struct {
	limits   ResourceLimits
	allowFS  []string // Allowed file paths
	allowAPI []string // Allowed API calls
}

// NewSandbox creates a new sandbox with the given limits.
func NewSandbox(limits ResourceLimits) *Sandbox {
	return &Sandbox{
		limits:   limits,
		allowFS:  []string{},
		allowAPI: []string{},
	}
}

// WithFileAccess adds allowed file paths to the sandbox.
func (s *Sandbox) WithFileAccess(paths ...string) *Sandbox {
	s.allowFS = append(s.allowFS, paths...)
	return s
}

// WithAPIAccess adds allowed API calls to the sandbox.
func (s *Sandbox) WithAPIAccess(apis ...string) *Sandbox {
	s.allowAPI = append(s.allowAPI, apis...)
	return s
}

// CheckFileAccess verifies if file access is allowed.
func (s *Sandbox) CheckFileAccess(path string) error {
	if len(s.allowFS) == 0 {
		// No restrictions if allowFS is empty
		return nil
	}

	for _, allowed := range s.allowFS {
		if path == allowed || matchesPattern(path, allowed) {
			return nil
		}
	}

	return fmt.Errorf("file access denied: %s", path)
}

// CheckAPIAccess verifies if API call is allowed.
func (s *Sandbox) CheckAPIAccess(api string) error {
	if len(s.allowAPI) == 0 {
		// No restrictions if allowAPI is empty
		return nil
	}

	for _, allowed := range s.allowAPI {
		if api == allowed || matchesPattern(api, allowed) {
			return nil
		}
	}

	return fmt.Errorf("API access denied: %s", api)
}

// GetLimits returns the sandbox resource limits.
func (s *Sandbox) GetLimits() ResourceLimits {
	return s.limits
}

// matchesPattern checks if a string matches a pattern (simple wildcard support).
func matchesPattern(s, pattern string) bool {
	// Simple pattern matching - can be enhanced with filepath.Match
	if pattern == "*" {
		return true
	}
	// For now, exact match
	return s == pattern
}
