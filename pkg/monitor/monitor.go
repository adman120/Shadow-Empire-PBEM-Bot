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

// FileTrackingInfo stores information about when a file was first seen and whether it has been processed.
type FileTrackingInfo struct {
	FirstSeen int64  // Timestamp when the file was first detected.
	Processed bool   // Flag indicating if the file has been processed.
}

// parseIgnorePatterns parses comma-separated ignore patterns from the environment variable IGNORE_PATTERNS.
// These patterns are used to determine which files should be ignored.
func parseIgnorePatterns() []string {
	patterns := os.Getenv("IGNORE_PATTERNS") // Get the value of the environment variable.
	if patterns == "" {
		return []string{} // Return an empty slice if the variable is not set.
	}
	var result []string
	for _, pattern := range strings.Split(patterns, ",") { // Split the string by commas.
		result = append(result, strings.ToLower(strings.TrimSpace(pattern))) // Trim spaces and convert to lowercase.
	}
	return result
}

// shouldIgnoreFile checks if a filename contains any of the ignore patterns.
// It returns true if the file should be ignored, false otherwise.
func shouldIgnoreFile(filename string, ignorePatterns []string) bool {
	if len(ignorePatterns) == 0 {
		return false // If there are no patterns, don't ignore any files.
	}

	lowerFilename := strings.ToLower(filename) // Convert the filename to lowercase for case-insensitive comparison.
	for _, pattern := range ignorePatterns {
		if strings.Contains(lowerFilename, pattern) {
			return true // If the filename contains any of the patterns, ignore it.
		}
	}
	return false
}

// MonitorDirectory monitors a directory for new save files and notifies the next player.
// This is the main function that starts the monitoring process.
func MonitorDirectory(dirPath string) {
	// Get username to Discord ID mappings from the environment variable USER_MAPPINGS.
	userMappings, err := userparser.ParseUsers("USER_MAPPINGS")
	if err != nil {
		log.Fatalf("❌ Failed to parse USER_MAPPINGS: %v. Please check the format (e.g., '1 User1 ID1,2 User2 ID2').", err)
	}

	// Parse ignore patterns from the environment variable.
	ignorePatterns := parseIgnorePatterns()
	if len(ignorePatterns) > 0 {
		fmt.Printf("🚫 Loaded %d ignore patterns\n", len(ignorePatterns))
	}

	// Log the parsed user mappings.  This is helpful for debugging.
	fmt.Printf("👥 Loaded %d user mappings:\n", len(userMappings))
	for _, mapping := range userMappings {
		fmt.Printf("  - Order: %d, User: %s, ID: %s\n", mapping.Order, mapping.Username, mapping.DiscordID)
	}

	// File tracking map with timestamps to implement debouncing.
	// The key is the filename (lowercase), and the value is a pointer to a FileTrackingInfo struct.
	fileTracker := make(map[string]*FileTrackingInfo)

	// Current turn tracking.  This is initialized to 1 and updated as new files are processed.
	currentTurn := 1

	// Get file debounce time from the environment or default to 30 seconds.
	// Debouncing is used to ensure that a file is completely written before it's processed.
	fileDebounceMs := 30000 // Default to 30000 milliseconds (30 seconds).
	if debounceEnv := os.Getenv("FILE_DEBOUNCE_MS"); debounceEnv != "" {
		if parsed, err := strconv.Atoi(debounceEnv); err == nil {
			fileDebounceMs = parsed // Override with the value from the environment variable if it's valid.
		}
	}
	fmt.Printf("⏱️ File debounce time set to %d seconds\n", fileDebounceMs/1000)

	// Initialize tracker with existing files as already processed.
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("❌ Error reading directory: %v\n", err)
		return // Exit the function if there's an error reading the directory.
	}

	for _, file := range files {
		if !file.IsDir() { // Only process files, not directories.
			lowerFilename := strings.ToLower(file.Name())
			fileTracker[lowerFilename] = &FileTrackingInfo{
				FirstSeen: time.Now().UnixMilli(), // Use the current time.
				Processed: true,                 // Mark existing files as processed.
			}
		}
	}
	fmt.Printf("📋 Initialized with %d existing files\n", len(fileTracker))

	// Set up polling interval (check every 5 seconds).
	pollInterval := 5 * time.Second

	fmt.Printf("👁️ Started monitoring directory: %s (polling every %v)\n", dirPath, pollInterval)

	ticker := time.NewTicker(pollInterval) // Create a ticker that ticks every pollInterval.
	defer ticker.Stop()                   // Ensure the ticker is stopped when the function exits.
	// Initialize lastCheckTime
	lastCheckTime := time.Now()

	for range ticker.C {
		currentTurn = processDirectory(dirPath, fileTracker, userMappings, fileDebounceMs, ignorePatterns, currentTurn, &lastCheckTime)
	}
}

// extractTurnNumber attempts to extract the turn number from a filename.
// It uses a regular expression to find the turn number in the filename.
func extractTurnNumber(filename string) int {
	// First try the standard pattern: something_turn#_something
	turnPattern := regexp.MustCompile(`_turn(\d+)_`) // Regular expression to find "_turn<number>_".
	matches := turnPattern.FindStringSubmatch(strings.ToLower(filename)) // Find the pattern in the lowercase filename.

	if len(matches) > 1 { // Check if there's a match (matches[0] is the whole string, matches[1] is the captured group).
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num // Convert the captured group to an integer and return it.
		}
	}

	return 0 // Return 0 if no turn number is found.
}

// processDirectory handles a single directory scan iteration.
// Returns the current turn number (possibly updated).
func processDirectory(dirPath string, fileTracker map[string]*FileTrackingInfo,
	userMappings []userparser.UserMapping,
	fileDebounceMs int, ignorePatterns []string, currentTurn int, lastCheckTime *time.Time) int {

	now := time.Now().UnixMilli()

	// Get the configured game name
	gameName := strings.ToLower(os.Getenv("GAME_NAME"))
	if gameName == "" {
		gameName = "pbem1"
	}

	// Track current files to detect deleted ones
	currentFiles := make(map[string]bool)
	var latestFileTime int64 = 0
	var latestFileName string

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

			// Find username in filename to identify the player whose turn it *is*
			var currentPlayerIndex = -1 // Index in the userMappings slice
			for i, mapping := range userMappings {
				// Check if the filename contains the *current* player's username (case-insensitive)
				if strings.Contains(filename, strings.ToLower(mapping.Username)) {
					currentPlayerIndex = i
					break
				}
			}

			if currentPlayerIndex != -1 {
				// The user found in the filename is the *current* player
				currentUserMapping := userMappings[currentPlayerIndex]

				// Determine the index of the *next* player in the order
				nextPlayerIndex := (currentPlayerIndex + 1) % len(userMappings)
				nextUserMapping := userMappings[nextPlayerIndex]

				// Determine the index of the player who just finished (previous player)
				previousPlayerIndex := (currentPlayerIndex - 1 + len(userMappings)) % len(userMappings)
				previousUserMapping := userMappings[previousUserIndex]

				// Determine the turn number for the *next* save file instruction
				saveInstructionTurnNumber := currentTurn
				// Check if the *current* player (whose file we are processing) is the last in the order.
				// If so, the save instruction should be for the *next* turn.
				if currentPlayerIndex == len(userMappings)-1 {
					saveInstructionTurnNumber = currentTurn + 1
					fmt.Printf("🔄 Last player (%s) finished turn %d, next save will start turn %d\n", currentUserMapping.Username, currentTurn, saveInstructionTurnNumber)
					// Update the main turn counter *after* processing this file and determining the instruction number
					currentTurn = saveInstructionTurnNumber
				}

				fmt.Printf("🔄 Turn %d: It's %s's turn (save from %s). Next up: %s (for turn %d)\n", currentTurn, currentUserMapping.Username, previousUserMapping.Username, nextUserMapping.Username, saveInstructionTurnNumber)

				// Send webhook to the *current* player, instructing them to save for the *next* player, using the correct turn number for the save instruction
				webhook.SendWebHook(currentUserMapping.Username, currentUserMapping.DiscordID, nextUserMapping.Username, saveInstructionTurnNumber)

				info.Processed = true
			} else {
				fmt.Printf("❓ Cannot match any user to save file: %s\n", filename)
				info.Processed = true
			}
		}
		//check for the latest file
		fileInfo, err := os.Stat(dirPath + "/" + file.Name())
		if err != nil {
			fmt.Printf("Error getting file info for %s: %v\n", file.Name(), err)
			continue
		}
		if fileInfo.ModTime().Unix() > latestFileTime {
			latestFileTime = fileInfo.ModTime().Unix()
			latestFileName = file.Name()
		}
	}
	checkFileAge(latestFileTime, latestFileName, lastCheckTime, userMappings)
	// Clean up tracking for deleted files
	for filename := range fileTracker {
		if !currentFiles[filename] {
			delete(fileTracker, filename)
			fmt.Printf("🗑️ Removed tracking for deleted file: %s\n", filename)
		}
	}

	return currentTurn
}

// checkFileAge checks the age of the latest file and sends a Discord notification if it exceeds the limit.
func checkFileAge(latestFileTime int64, latestFileName string, lastCheckTime *time.Time, userMappings []userparser.UserMapping) {
	fileCheckTime := 24 * time.Hour //check every 24 hours
	fileCheckTimeEnv := os.Getenv("FILE_CHECK_TIME")
	if fileCheckTimeEnv != "" {
		duration, err := time.ParseDuration(fileCheckTimeEnv)
		if err != nil {
			log.Printf("Invalid FILE_CHECK_TIME value: %s. Using default (24h).\n", fileCheckTimeEnv)
		} else {
			fileCheckTime = duration
		}
	}

	now := time.Now()
	if now.Sub(*lastCheckTime) >= fileCheckTime {
		*lastCheckTime = now
		fileAgeLimit := 24 * time.Hour // Default to 24 hours
		fileAgeLimitEnv := os.Getenv("FILE_AGE_LIMIT")
		if fileAgeLimitEnv != "" {
			duration, err := time.ParseDuration(fileAgeLimitEnv)
			if err != nil {
				log.Printf("Invalid FILE_AGE_LIMIT value: %s. Using default (24h).\n", fileAgeLimitEnv)
			} else {
				fileAgeLimit = duration
			}
		}

		fileAge := time.Duration(now.Unix() - latestFileTime) * time.Second

		if fileAge > fileAgeLimit {
			fmt.Printf("⏰ Latest file (%s) is older than %v (%v old). Sending Discord notification.\n", latestFileName, fileAgeLimit, fileAge)
			//find user
			var currentPlayerIndex = -1 // Index in the userMappings slice
			for i, mapping := range userMappings {
				// Check if the filename contains the  player's username (case-insensitive)
				if strings.Contains(latestFileName, strings.ToLower(mapping.Username)) {
					currentPlayerIndex = i
					break
				}
			}
			if currentPlayerIndex != -1 {
				currentUserMapping := userMappings[currentPlayerIndex]
				webhook.SendFileAgeWarningWebHook(latestFileName, fileAge, fileAgeLimit, currentUserMapping.Username, currentUserMapping.DiscordID)
			} else {
				webhook.SendFileAgeWarningWebHook(latestFileName, fileAge, fileAgeLimit, "", "")
			}

		} else {
			fmt.Printf("Latest file (%s) is %v old, which is within the limit (%v).\n", latestFileName, fileAge, fileAgeLimit)
		}
	}
}

