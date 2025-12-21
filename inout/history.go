package inout

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/scottyeager/pal/config"
)

// History manages conversation history storage and retrieval
type History struct {
	basePath string
}

// Thread represents a conversation thread
type Thread struct {
	ID           int
	Date         time.Time
	CommandName  string // Store command name
	SystemPrompt string // Store actual system prompt content
	Turns        []Turn
	Path         string
}

// Turn represents a single exchange (prompt + response) in a thread
type Turn struct {
	ID       int
	Prompt   string
	Response string
}

// NewHistory creates a new History instance
func NewHistory() (*History, error) {
	basePath, err := getHistoryBasePath()
	if err != nil {
		return nil, err
	}

	// Ensure history directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	return &History{
		basePath: basePath,
	}, nil
}

// getHistoryBasePath returns the base path for history storage
func getHistoryBasePath() (string, error) {
	basePath, err := config.GetBasePath()
	if err != nil {
		return "", fmt.Errorf("failed to get base path: %w", err)
	}
	return filepath.Join(basePath, "history"), nil
}

// getDatePath returns the path for a specific date's threads
func (h *History) getDatePath(date time.Time) string {
	return filepath.Join(h.basePath, date.Format("2006-01-02"))
}

// getThreadPath returns the path for a specific thread
func (h *History) getThreadPath(date time.Time, threadID int) string {
	datePath := h.getDatePath(date)
	return filepath.Join(datePath, strconv.Itoa(threadID))
}

// getTurnPath returns the path for a specific turn file
func (h *History) getTurnPath(date time.Time, threadID int, turnID int, isPrompt bool) string {
	threadPath := h.getThreadPath(date, threadID)
	filename := strconv.Itoa(turnID)
	if isPrompt {
		filename += "-prompt"
	} else {
		filename += "-response"
	}
	return filepath.Join(threadPath, filename)
}

// StartNewThread starts a new conversation thread for today
func (h *History) StartNewThread(commandName string) (*Thread, error) {
	today := time.Now()
	datePath := h.getDatePath(today)

	// Ensure date directory exists
	if err := os.MkdirAll(datePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create date directory: %w", err)
	}

	// Find the next available thread ID
	threadID := 0
	entries, err := os.ReadDir(datePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read date directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if id, err := strconv.Atoi(entry.Name()); err == nil {
				if id >= threadID {
					threadID = id + 1
				}
			}
		}
	}

	// Create thread directory
	threadPath := h.getThreadPath(today, threadID)
	if err := os.MkdirAll(threadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create thread directory: %w", err)
	}

	// Store command name in 0-{commandName} file (content can be empty initially)
	commandPath := filepath.Join(threadPath, "0-"+commandName)
	if err := os.WriteFile(commandPath, []byte(""), 0644); err != nil {
		return nil, fmt.Errorf("failed to write command file: %w", err)
	}

	return &Thread{
		ID:           threadID,
		Date:         today,
		CommandName:  commandName,
		SystemPrompt: "", // Will be set by UpdateSystemPrompt if needed
		Turns:        []Turn{},
		Path:         threadPath,
	}, nil
}

// GetOrCreateThread gets an existing thread or creates a new one
// If threadID is -1, creates a new thread
// If threadID is >= 0, tries to load that thread
func (h *History) GetOrCreateThread(threadID int, dateStr string) (*Thread, error) {
	var date time.Time
	var err error

	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	if threadID == -1 {
		return h.StartNewThread("")
	}

	// Try to load existing thread
	return h.loadThread(date, threadID)
}

// loadThread loads an existing thread from disk
func (h *History) loadThread(date time.Time, threadID int) (*Thread, error) {
	threadPath := h.getThreadPath(date, threadID)

	// Check if thread directory exists
	if _, err := os.Stat(threadPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("thread %d not found for date %s", threadID, date.Format("2006-01-02"))
	}

	// Read command name and system prompt from 0-{commandName} file
	commandName := ""
	systemPrompt := ""

	entries, err := os.ReadDir(threadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read thread directory: %w", err)
	}

	// Look for 0-{commandName} file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if strings.HasPrefix(filename, "0-") {
			commandName = strings.TrimPrefix(filename, "0-")
			// Read the file content as system prompt
			filePath := filepath.Join(threadPath, filename)
			if content, err := os.ReadFile(filePath); err == nil {
				systemPrompt = string(content)
			}
			break
		}
	}

	turns := []Turn{}
	turnMap := make(map[int]*Turn)

	// First pass: collect all turn files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Skip command file
		if strings.HasPrefix(filename, "0-") {
			continue
		}

		var turnID int
		var isPrompt bool

		if strings.HasSuffix(filename, "-prompt") {
			idStr := strings.TrimSuffix(filename, "-prompt")
			if id, err := strconv.Atoi(idStr); err == nil {
				turnID = id
				isPrompt = true
			} else {
				continue // Skip invalid files
			}
		} else if strings.HasSuffix(filename, "-response") {
			idStr := strings.TrimSuffix(filename, "-response")
			if id, err := strconv.Atoi(idStr); err == nil {
				turnID = id
				isPrompt = false
			} else {
				continue // Skip invalid files
			}
		} else {
			continue // Skip files that don't match pattern
		}

		// Get or create turn
		turn, exists := turnMap[turnID]
		if !exists {
			turn = &Turn{ID: turnID}
			turnMap[turnID] = turn
		}

		// Read file content
		filePath := filepath.Join(threadPath, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read turn file %s: %w", filename, err)
		}

		if isPrompt {
			turn.Prompt = string(content)
		} else {
			turn.Response = string(content)
		}
	}

	// Convert map to sorted slice
	// Get all keys and sort them
	keys := make([]int, 0, len(turnMap))
	for k := range turnMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// Add turns in sorted order
	for _, k := range keys {
		turns = append(turns, *turnMap[k])
	}

	return &Thread{
		ID:           threadID,
		Date:         date,
		CommandName:  commandName,
		SystemPrompt: systemPrompt,
		Turns:        turns,
		Path:         threadPath,
	}, nil
}

// AddTurn adds a new turn (prompt + response) to a thread
func (h *History) AddTurn(thread *Thread, prompt, response string) error {
	turnID := len(thread.Turns) + 1

	// Write prompt file
	promptPath := h.getTurnPath(thread.Date, thread.ID, turnID, true)
	if err := os.WriteFile(promptPath, []byte(prompt), 0644); err != nil {
		return fmt.Errorf("failed to write prompt: %w", err)
	}

	// Write response file
	responsePath := h.getTurnPath(thread.Date, thread.ID, turnID, false)
	if err := os.WriteFile(responsePath, []byte(response), 0644); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	// Update thread in memory
	thread.Turns = append(thread.Turns, Turn{
		ID:       turnID,
		Prompt:   prompt,
		Response: response,
	})

	return nil
}

// GetLatestThread gets the most recent thread
func (h *History) GetLatestThread() (*Thread, error) {
	today := time.Now()

	// Try today's threads first
	datePath := h.getDatePath(today)
	if entries, err := os.ReadDir(datePath); err == nil {
		// Find highest thread ID
		maxID := -1
		for _, entry := range entries {
			if entry.IsDir() {
				if id, err := strconv.Atoi(entry.Name()); err == nil && id > maxID {
					maxID = id
				}
			}
		}

		if maxID >= 0 {
			return h.loadThread(today, maxID)
		}
	}

	// If no threads today, look for yesterday
	yesterday := today.AddDate(0, 0, -1)
	datePath = h.getDatePath(yesterday)
	if entries, err := os.ReadDir(datePath); err == nil {
		// Find highest thread ID
		maxID := -1
		for _, entry := range entries {
			if entry.IsDir() {
				if id, err := strconv.Atoi(entry.Name()); err == nil && id > maxID {
					maxID = id
				}
			}
		}

		if maxID >= 0 {
			return h.loadThread(yesterday, maxID)
		}
	}

	// No threads found, create a new one
	return h.StartNewThread("")
}

// ListThreads lists all threads for a given date (or today if empty)
func (h *History) ListThreads(dateStr string) ([]Thread, error) {
	var date time.Time
	var err error

	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	datePath := h.getDatePath(date)
	entries, err := os.ReadDir(datePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Thread{}, nil
		}
		return nil, fmt.Errorf("failed to read date directory: %w", err)
	}

	var threads []Thread
	for _, entry := range entries {
		if entry.IsDir() {
			if id, err := strconv.Atoi(entry.Name()); err == nil {
				thread, err := h.loadThread(date, id)
				if err != nil {
					// Skip threads that can't be loaded
					continue
				}
				threads = append(threads, *thread)
			}
		}
	}

	return threads, nil
}

// GetThreadByPath gets a thread by its path string (e.g., "2025-12-12/0")
func (h *History) GetThreadByPath(path string) (*Thread, error) {
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid thread path format: expected date/threadID")
	}

	dateStr := parts[0]
	threadIDStr := parts[1]

	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid thread ID: %w", err)
	}

	return h.GetOrCreateThread(threadID, dateStr)
}

// UpdateSystemPrompt updates the system prompt content for a thread
func (h *History) UpdateSystemPrompt(thread *Thread, systemPrompt string) error {
	commandPath := filepath.Join(thread.Path, "0-"+thread.CommandName)
	if err := os.WriteFile(commandPath, []byte(systemPrompt), 0644); err != nil {
		return fmt.Errorf("failed to update system prompt: %w", err)
	}
	thread.SystemPrompt = systemPrompt
	return nil
}

// CopyThread creates a copy of a thread with a new ID
func (h *History) CopyThread(source *Thread) (*Thread, error) {
	// Create new thread with the same command name
	newThread, err := h.StartNewThread(source.CommandName)
	if err != nil {
		return nil, err
	}

	// Copy system prompt if it exists
	if source.SystemPrompt != "" {
		if err := h.UpdateSystemPrompt(newThread, source.SystemPrompt); err != nil {
			return nil, err
		}
	}

	// Copy all turns
	for _, turn := range source.Turns {
		if err := h.AddTurn(newThread, turn.Prompt, turn.Response); err != nil {
			return nil, err
		}
	}

	return newThread, nil
}
