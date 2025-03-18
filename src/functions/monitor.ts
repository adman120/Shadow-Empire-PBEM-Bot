import * as fs from "fs";

import { sendWebHook } from "./webhook";
import { userparse } from "./userParser";

// File tracking interface to store when a file was first seen
interface FileTrackingInfo {
  firstSeen: number;
  processed: boolean;
}

/**
 * Parse comma-separated ignore patterns from environment variable
 */
function parseIgnorePatterns(): string[] {
  const patterns = process.env.IGNORE_PATTERNS;
  if (!patterns) {
    return [];
  }
  return patterns.split(',').map(pattern => pattern.trim().toLowerCase());
}

/**
 * Check if a filename contains any of the ignore patterns
 */
function shouldIgnoreFile(filename: string, ignorePatterns: string[]): boolean {
  if (ignorePatterns.length === 0) return false;
  
  const lowerFilename = filename.toLowerCase();
  return ignorePatterns.some(pattern => lowerFilename.includes(pattern));
}

/**
 * Monitor a directory for new save files and notify the next player in the turn order
 */
export function monitorDirectory(dirPath: string): void {
  // Get username to Discord ID mappings from environment variable
  const usernameToDiscordId: Record<string, string> =
    userparse("USER_MAPPINGS");

  // Parse ignore patterns
  const ignorePatterns = parseIgnorePatterns();
  if (ignorePatterns.length > 0) {
    console.log(`🚫 Loaded ${ignorePatterns.length} ignore patterns`);
  }

  // Log the parsed user mappings
  console.log(`👥 Loaded ${Object.keys(usernameToDiscordId).length} user mappings`);
  
  const userList = Object.keys(usernameToDiscordId);

  // File tracking map with timestamps to implement debouncing
  const fileTracker: Map<string, FileTrackingInfo> = new Map();
  
  // Get file debounce time from environment or default to 30 seconds
  const fileDebounceMs = parseInt(process.env.FILE_DEBOUNCE_MS || '30000', 10);
  console.log(`⏱️ File debounce time set to ${fileDebounceMs/1000} seconds`);

  // Initialize tracker with existing files as already processed
  fs.readdirSync(dirPath).forEach(file => {
    const lowerFilename = file.toLowerCase();
    fileTracker.set(lowerFilename, { firstSeen: Date.now(), processed: true });
  });
  console.log(`📋 Initialized with ${fileTracker.size} existing files`);

  // Set up polling interval (check every 5 seconds)
  const POLL_INTERVAL = 5000; // 5 seconds
  
  setInterval(() => {
    try {
      const now = Date.now();
      
      // Get current files in the directory
      const currentFiles = fs.readdirSync(dirPath).map(file => file.toLowerCase());
      const currentFilesSet = new Set(currentFiles);
      
      // Check for new files or files ready for processing
      for (const file of currentFiles) {
        if (!fileTracker.has(file)) {
          // New file detected, start tracking it
          console.log(`📄 New save file detected: ${file}, starting debounce period`);
          fileTracker.set(file, { firstSeen: now, processed: false });
        } else {
          // File exists in tracker, check if it's ready for processing
          const fileInfo = fileTracker.get(file)!;
          
          if (!fileInfo.processed && (now - fileInfo.firstSeen) >= fileDebounceMs) {
            // File has been stable for the debounce period, process it
            console.log(`⏱️ File ${file} stable for ${fileDebounceMs/1000}s, processing now`);
            
            // Check if the file should be ignored based on patterns
            if (shouldIgnoreFile(file, ignorePatterns)) {
              console.log(`🚫 Ignoring file ${file} based on ignore patterns`);
              fileInfo.processed = true;
              continue;
            }
            
            // Check for the username inside the filename
            const username = userList.find((name) =>
              file.includes(name.toLowerCase())
            );
            
            if (username) {
              const discordId = usernameToDiscordId[username];
              
              // Find the index of current user and determine next user
              const currentUserIndex = userList.indexOf(username);
              const nextUserIndex = (currentUserIndex + 1) % userList.length;
              const nextUser = userList[nextUserIndex];
              
              console.log(`🔄 Turn passing from ${username} to ${nextUser}`);
              sendWebHook(username, discordId, nextUser);
              
              // Mark as processed
              fileInfo.processed = true;
            } else {
              console.log(`❓ Cannot match any user to save file: ${file}`);
              fileInfo.processed = true;
            }
          }
        }
      }
      
      // Clean up tracking for files that no longer exist
      for (const [trackedFile] of fileTracker) {
        if (!currentFilesSet.has(trackedFile)) {
          fileTracker.delete(trackedFile);
          console.log(`🗑️ Removed tracking for deleted file: ${trackedFile}`);
        }
      }
    } catch (error) {
      console.error(`❌ Error polling directory: ${error}`);
    }
  }, POLL_INTERVAL);
  
  console.log(`👁️ Started monitoring directory: ${dirPath} (polling every ${POLL_INTERVAL/1000}s)`);
}
