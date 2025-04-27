package monitor

import (
	"fmt"
	"log"
	"os"
	"regexp"
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
	userMappings, err := userparser.ParseUsers("USER_MAPPINGS")
	if err != nil {
		log.Fatalf("❌ Failed to parse USER_MAPPINGS: %v. Please check the format (e.g., '1 User1 ID1,2 User2 ID2').", err)
	}

	// Parse ignore patterns
	ignorePatterns := parseIgnorePatterns()
	if len(ignorePatterns) > 0 {
		fmt.Printf("🚫 Loaded %d ignore patterns\n", len(ignorePatterns))
	}

	// Log the parsed user mappings
	fmt.Printf("👥 Loaded %d user mappings:\n", len(userMappings))
	for _, mapping := range userMappings {
		fmt.Printf("  - Order: %d, User: %s, ID: %s\n", mapping.Order, mapping.Username, mapping.DiscordID)
	}

	// File tracking map with timestamps to implement debouncing
	fileTracker := make(map[string]*FileTrackingInfo)

	// Current turn tracking
	currentTurn := 1

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
		currentTurn = processDirectory(dirPath, fileTracker, userMappings, fileDebounceMs, ignorePatterns, currentTurn)
	}
}

// extractTurnNumber attempts to extract turn number from a filename
func extractTurnNumber(filename string) int {
	// First try the standard pattern: something_turn#_something
	turnPattern := regexp.MustCompile(`_turn(\d+)_`)
	matches := turnPattern.FindStringSubmatch(strings.ToLower(filename))

	if len(matches) > 1 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num
		}
	}

	return 0 // Return 0 if no turn number found
}

// processDirectory handles a single directory scan iteration
// Returns the current turn number (possibly updated)
func processDirectory(dirPath string, fileTracker map[string]*FileTrackingInfo,
	userMappings []userparser.UserMapping,
	fileDebounceMs int, ignorePatterns []string, currentTurn int) int {

	now := time.Now().UnixMilli()

	// Get the configured game name
	gameName := strings.ToLower(os.Getenv("GAME_NAME"))
	if gameName == "" {
		gameName = "pbem1"
	}

	// Track current files to detect deleted ones
	currentFiles := make(map[string]bool)

	// Read all files in directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("❌ Error reading directory: %v\n", err)
		return currentTurn
	}

	// Process each file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := strings.ToLower(file.Name())
		currentFiles[filename] = true

		// Try to extract turn number from filename
		if turnNumber := extractTurnNumber(filename); turnNumber > currentTurn {
			currentTurn = turnNumber
			fmt.Printf("🔢 Updated current turn to %d based on filename: %s\n", currentTurn, filename)
		}

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

			// Check if the game name in the filename matches the configured game name
			if !strings.HasPrefix(filename, gameName) {
				fmt.Printf("⚠️ File %s doesn't match configured game name '%s'\n", filename, gameName)

				// Try to find which user *might* have saved this based on filename content
				var foundUserIndex = -1 // Index in the userMappings slice
				for i, mapping := range userMappings {
					if strings.Contains(filename, strings.ToLower(mapping.Username)) {
						foundUserIndex = i
						break
					}
				}

				// Find the previous user who should be notified about the naming issue
				if foundUserIndex != -1 {
					// Determine the index of the user who *should* have saved (previous user in order)
					previousUserIndex := (foundUserIndex - 1 + len(userMappings)) % len(userMappings)
					previousUserMapping := userMappings[previousUserIndex]

					fmt.Printf("🔔 Sending rename notification to previous user %s (%s) for incorrectly named file %s\n",
						previousUserMapping.Username, previousUserMapping.DiscordID, filename)
					webhook.SendRenameWebHook(previousUserMapping.Username, previousUserMapping.DiscordID, filename, currentTurn)

				} else {
					fmt.Printf("❓ Cannot identify any user for incorrectly named file: %s. Cannot determine who to notify.\n", filename)
				}

				info.Processed = true
				continue
			}

			// Find username in filename to identify the *next* player
			var foundUserIndex = -1 // Index in the userMappings slice
			for i, mapping := range userMappings {
				// Check if the filename contains the *next* player's username (case-insensitive)
				// The save file format is typically game_turnX_nextPlayer
				if strings.Contains(filename, strings.ToLower(mapping.Username)) {
					foundUserIndex = i
					break
				}
			}

			if foundUserIndex != -1 {
				// The user found in the filename is the *next* player
				nextUserMapping := userMappings[foundUserIndex]

				// Determine the index of the user who just finished their turn (previous user in order)
				currentUserIndex := (foundUserIndex - 1 + len(userMappings)) % len(userMappings)
				currentUserMapping := userMappings[currentUserIndex]

				// If the next user is the first in the list, increment the turn
				// (assuming the list is sorted 1, 2, 3...)
				if nextUserMapping.Order == 1 { // Check if the *next* user is the first one
					currentTurn++
					fmt.Printf("🔄 Full player rotation completed, incrementing turn to %d\n", currentTurn)
				}

				fmt.Printf("🔄 Turn %d passing from %s to %s\n", currentTurn, currentUserMapping.Username, nextUserMapping.Username)
				// Send webhook to the *next* player
				webhook.SendWebHook(currentUserMapping.Username, nextUserMapping.DiscordID, nextUserMapping.Username, currentTurn)

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

	return currentTurn
}
