// Package persistence provides session marker storage abstractions.
// This decouples session management from filesystem operations,
// improving testability and enabling alternative storage backends.
package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// SessionMarkerStore defines the interface for session marker persistence.
// Session markers are used to track whether a CLI session can be resumed
// (e.g., Claude Code's --resume functionality).
type SessionMarkerStore interface {
	// Exists checks if a session marker exists for the given session ID.
	Exists(sessionID string) bool

	// Create creates a session marker for the given session ID.
	// Returns an error if the marker cannot be created.
	Create(sessionID string) error

	// Delete removes the session marker for the given session ID.
	// Returns an error if the marker cannot be deleted.
	Delete(sessionID string) error

	// Dir returns the base directory where markers are stored.
	Dir() string
}

// FileMarkerStore implements SessionMarkerStore using the local filesystem.
// Markers are stored as empty files with a .lock extension.
type FileMarkerStore struct {
	markerDir string
	mu        sync.RWMutex
}

// NewFileMarkerStore creates a new FileMarkerStore with the given base directory.
// If the directory doesn't exist, it will be created.
func NewFileMarkerStore(markerDir string) (*FileMarkerStore, error) {
	if markerDir == "" {
		return nil, fmt.Errorf("marker directory cannot be empty")
	}

	// Create the marker directory if it doesn't exist
	if err := os.MkdirAll(markerDir, 0755); err != nil {
		return nil, fmt.Errorf("create marker directory: %w", err)
	}

	return &FileMarkerStore{
		markerDir: markerDir,
	}, nil
}

// NewDefaultFileMarkerStore creates a FileMarkerStore in the default location.
// Uses ~/.hotplex/sessions on success, falls back to temp directory on failure.
func NewDefaultFileMarkerStore() *FileMarkerStore {
	var markerDir string

	homeDir, err := os.UserHomeDir()
	if err == nil {
		markerDir = filepath.Join(homeDir, ".hotplex", "sessions")
	} else {
		markerDir = filepath.Join(os.TempDir(), "hotplex_sessions")
	}

	// Best effort creation
	_ = os.MkdirAll(markerDir, 0755)

	return &FileMarkerStore{
		markerDir: markerDir,
	}
}

// Exists checks if a session marker file exists.
func (s *FileMarkerStore) Exists(sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	markerPath := s.markerPath(sessionID)
	_, err := os.Stat(markerPath)
	return err == nil
}

// Create creates a session marker file.
func (s *FileMarkerStore) Create(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	markerPath := s.markerPath(sessionID)
	if err := os.WriteFile(markerPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("create session marker: %w", err)
	}
	return nil
}

// Delete removes a session marker file.
func (s *FileMarkerStore) Delete(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	markerPath := s.markerPath(sessionID)
	if err := os.Remove(markerPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete session marker: %w", err)
	}
	return nil
}

// Dir returns the marker directory path.
func (s *FileMarkerStore) Dir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.markerDir
}

// markerPath returns the full path to a session marker file.
func (s *FileMarkerStore) markerPath(sessionID string) string {
	return filepath.Join(s.markerDir, sessionID+".lock")
}

// InMemoryMarkerStore implements SessionMarkerStore using an in-memory map.
// Useful for testing and scenarios where persistence is not required.
type InMemoryMarkerStore struct {
	markers map[string]bool
	mu      sync.RWMutex
}

// NewInMemoryMarkerStore creates a new InMemoryMarkerStore.
func NewInMemoryMarkerStore() *InMemoryMarkerStore {
	return &InMemoryMarkerStore{
		markers: make(map[string]bool),
	}
}

// Exists checks if a session marker exists in memory.
func (s *InMemoryMarkerStore) Exists(sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.markers[sessionID]
}

// Create creates a session marker in memory.
func (s *InMemoryMarkerStore) Create(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.markers[sessionID] = true
	return nil
}

// Delete removes a session marker from memory.
func (s *InMemoryMarkerStore) Delete(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.markers, sessionID)
	return nil
}

// Dir returns an empty string for in-memory store.
func (s *InMemoryMarkerStore) Dir() string {
	return ""
}
