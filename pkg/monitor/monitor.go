package monitor

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/1Solon/shadow-empire-pbem-bot/pkg/userparser"
	"github.com/1Solon/shadow-empire-pbem-bot/pkg/webhook"
)

// FileTrackingInfo stores information about when a file was first seen
type FileTrackingInfo struct {
	FirstSeen int64
	Processed bool
}

// parseIgnorePatterns parses comma-separated ignore patterns from environment variable
func parseIgnorePatterns() []string {
	patterns := os.Getenv("IGNORE_PATTERNS")
	if patterns == "" {
		return []string{}
	}
	var result []string
	for _, pattern := range strings.Split(patterns, ",") {
		result = append(result, strings.ToLower(strings.TrimSpace(pattern)))
	}
	return result
}

// shouldIgnoreFile checks if a filename contains any of the ignore patterns
func shouldIgnoreFile(filename string, ignorePatterns []string) bool {
	if len(ignorePatterns) == 0 {
		return false
	}

	lowerFilename := strings.ToLower(filename)
	for _, pattern := range ignorePatterns {
		if strings.Contains(lowerFilename, pattern) {
			return true
		}
	}
	return false
}

// MonitorDirectory monitors a directory for new save files and notifies the next player
func MonitorDirectory(dirPath string) {
	// Get username to Discord ID mappings from environment variable
	usernameToDiscordID := userparser.ParseUsers("USER_MAPPINGS")

	// Parse ignore patterns
	ignorePatterns := parseIgnorePatterns()
	if len(ignorePatterns) > 0 {
		fmt.Printf("🚫 Loaded %d ignore patterns\n", len(ignorePatterns))
	}

	// Log the parsed user mappings
	fmt.Printf("👥 Loaded %d user mappings\n", len(usernameToDiscordID))

	// Extract user list from map keys
	var userList []string
	for username := range usernameToDiscordID {
		userList = append(userList, username)
	}

	// File tracking map with timestamps to implement debouncing
	fileTracker := make(map[string]*FileTrackingInfo)

	// Get file debounce time from environment or default to 30 seconds
	fileDebounceMs := 30000
	if debounceEnv := os.Getenv("FILE_DEBOUNCE_MS"); debounceEnv != "" {
		if parsed, err := strconv.Atoi(debounceEnv); err == nil {
			fileDebounceMs = parsed
		}
	}
	fmt.Printf("⏱️ File debounce time set to %d seconds\n", fileDebounceMs/1000)

	// Initialize tracker with existing files as already processed
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("❌ Error reading directory: %v\n", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			lowerFilename := strings.ToLower(file.Name())
			fileTracker[lowerFilename] = &FileTrackingInfo{
				FirstSeen: time.Now().UnixMilli(),
				Processed: true,
			}
		}
	}
	fmt.Printf("📋 Initialized with %d existing files\n", len(fileTracker))

	// Set up polling interval (check every 5 seconds)
	pollInterval := 5 * time.Second

	fmt.Printf("👁️ Started monitoring directory: %s (polling every %v)\n", dirPath, pollInterval)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		processDirectory(dirPath, fileTracker, usernameToDiscordID, userList, fileDebounceMs, ignorePatterns)
	}
}

// processDirectory handles a single directory scan iteration
func processDirectory(dirPath string, fileTracker map[string]*FileTrackingInfo,
	usernameToDiscordID map[string]string, userList []string,
	fileDebounceMs int, ignorePatterns []string) {

	now := time.Now().UnixMilli()

	// Track current files to detect deleted ones
	currentFiles := make(map[string]bool)

	// Read all files in directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("❌ Error reading directory: %v\n", err)
		return
	}

	// Process each file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := strings.ToLower(file.Name())
		currentFiles[filename] = true

		if info, exists := fileTracker[filename]; !exists {
			// New file detected
			fmt.Printf("📄 New save file detected: %s, starting debounce period\n", filename)
			fileTracker[filename] = &FileTrackingInfo{
				FirstSeen: now,
				Processed: false,
			}
		} else if !info.Processed && (now-info.FirstSeen) >= int64(fileDebounceMs) {
			// File has been stable for debounce period
			fmt.Printf("⏱️ File %s stable for %ds, processing now\n", filename, fileDebounceMs/1000)

			// Check if the file should be ignored
			if shouldIgnoreFile(filename, ignorePatterns) {
				fmt.Printf("🚫 Ignoring file %s based on ignore patterns\n", filename)
				info.Processed = true
				continue
			}

			// Find username in filename
			var foundUser string
			for _, username := range userList {
				if strings.Contains(filename, strings.ToLower(username)) {
					foundUser = username
					break
				}
			}

			if foundUser != "" {
				discordID := usernameToDiscordID[foundUser]

				// Find next user
				nextUserIndex := -1
				for i, user := range userList {
					if user == foundUser {
						nextUserIndex = (i + 1) % len(userList)
						break
					}
				}

				if nextUserIndex >= 0 {
					nextUser := userList[nextUserIndex]
					fmt.Printf("🔄 Turn passing from %s to %s\n", foundUser, nextUser)
					webhook.SendWebHook(foundUser, discordID, nextUser)
				}

				info.Processed = true
			} else {
				fmt.Printf("❓ Cannot match any user to save file: %s\n", filename)
				info.Processed = true
			}
		}
	}

	// Clean up tracking for deleted files
	for filename := range fileTracker {
		if !currentFiles[filename] {
			delete(fileTracker, filename)
			fmt.Printf("🗑️ Removed tracking for deleted file: %s\n", filename)
		}
	}
}
