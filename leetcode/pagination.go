package leetcode

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type PageData struct {
	Problems   []any
	Page       int
	PageSize   int
	TotalPages int
	Metadata   map[string]string
	ExpiresAt  time.Time
}

type PaginationManager struct {
	mu         sync.RWMutex
	pages      map[string]*PageData
	defaultTTL time.Duration
}

// NewPaginationManager creates a manager with configurable session TTL
func NewPaginationManager(ttl time.Duration) *PaginationManager {
	pm := &PaginationManager{
		pages:      make(map[string]*PageData),
		defaultTTL: ttl,
	}
	// Start background cleanup goroutine
	go pm.startCleanupLoop()
	return pm
}

// Store saves pagination state for a user
func (pm *PaginationManager) Store(userID, command string, items []any, pageSize int, metadata map[string]string) {
	totalPages := (len(items) + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.pages[userID+"|"+command] = &PageData{
		Problems:   items,
		Page:       0,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Metadata:   metadata,
		ExpiresAt:  time.Now().Add(pm.defaultTTL),
	}
}

// Get retrieves pagination state (returns nil if expired/missing)
func (pm *PaginationManager) Get(userID, command string) *PageData {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	data, ok := pm.pages[userID+"|"+command]
	if !ok || time.Now().After(data.ExpiresAt) {
		if ok {
			// Clean up expired entry
			pm.mu.RUnlock()
			pm.mu.Lock()
			delete(pm.pages, userID+"|"+command)
			pm.mu.Unlock()
			pm.mu.RLock()
		}
		return nil
	}
	return data
}

// Navigate updates the page number with bounds checking
// Returns the updated PageData or an error if session invalid
func (pm *PaginationManager) Navigate(userID, command string, delta int) (*PageData, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	key := userID + "|" + command
	data, ok := pm.pages[key]
	if !ok {
		return nil, fmt.Errorf("pagination session not found")
	}
	if time.Now().After(data.ExpiresAt) {
		delete(pm.pages, key)
		return nil, fmt.Errorf("session expired")
	}

	newPage := data.Page + delta
	if newPage < 0 {
		newPage = 0
	} else if newPage >= data.TotalPages {
		newPage = data.TotalPages - 1
	}
	data.Page = newPage
	return data, nil
}

// BuildButtonID creates a consistent CustomID for pagination buttons
func BuildButtonID(command string, action, page int, metadata ...string) string {
	parts := []string{command, fmt.Sprintf("%d", action), fmt.Sprintf("%d", page)}
	parts = append(parts, metadata...)
	return strings.Join(parts, "|")
}

// ParseButtonID extracts components from a CustomID
func ParseButtonID(customID string) (command string, action int, page int, metadata []string, err error) {
	parts := strings.Split(customID, "|")
	if len(parts) < 3 {
		return "", 0, 0, nil, fmt.Errorf("invalid button ID format")
	}

	command = parts[0]
	fmt.Sscanf(parts[1], "%d", &action)
	fmt.Sscanf(parts[2], "%d", &page)
	if len(parts) > 3 {
		metadata = parts[3:]
	}
	return command, action, page, metadata, nil
}

// BuildPaginationButtons creates standard Prev/Next buttons
func BuildPaginationButtons(command string, currentPage, totalPages int, metadata ...string) []discordgo.MessageComponent {
	prevDisabled := currentPage == 0
	nextDisabled := currentPage >= totalPages-1

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					Label:    "◀ Prev",
					Style:    discordgo.PrimaryButton,
					CustomID: BuildButtonID(command, -1, currentPage, metadata...),
					Disabled: prevDisabled,
				},
				&discordgo.Button{
					Label:    "Next ▶",
					Style:    discordgo.PrimaryButton,
					CustomID: BuildButtonID(command, 1, currentPage, metadata...),
					Disabled: nextDisabled,
				},
			},
		},
	}
}

// startCleanupLoop periodically removes expired sessions
func (pm *PaginationManager) startCleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		pm.mu.Lock()
		now := time.Now()
		for key, data := range pm.pages {
			if now.After(data.ExpiresAt) {
				delete(pm.pages, key)
			}
		}
		pm.mu.Unlock()
	}
}

// Global instance (created in main.go)
var DefaultPagination *PaginationManager

// InitPagination must be called once at startup
func InitPagination(ttl time.Duration) {
	DefaultPagination = NewPaginationManager(ttl)
}
