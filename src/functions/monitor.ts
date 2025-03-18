import * as fs from "fs";

import { sendWebHook } from "./webhook";
import { userparse } from "./userParser";

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

  // Initialize set of known files
  const knownFiles: Set<string> = new Set(
    fs.readdirSync(dirPath).map(file => file.toLowerCase())
  );
  console.log(`📋 Initialized with ${knownFiles.size} existing files`);

  // Set up polling interval (check every 5 seconds)
  const POLL_INTERVAL = 5000; // 5 seconds
  
  setInterval(() => {
    try {
      // Get current files in the directory
      const currentFiles = fs.readdirSync(dirPath).map(file => file.toLowerCase());
      
      // Check for new files
      for (const file of currentFiles) {
        if (!knownFiles.has(file)) {
          console.log(`📄 New save file detected: ${file}`);
          knownFiles.add(file);
          
          // Check if the file should be ignored based on patterns
          if (shouldIgnoreFile(file, ignorePatterns)) {
            console.log(`🚫 Ignoring file ${file} based on ignore patterns`);
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
          } else {
            console.log(`❓ Cannot match any user to save file: ${file}`);
          }
        }
      }
    } catch (error) {
      console.error(`❌ Error polling directory: ${error}`);
    }
  }, POLL_INTERVAL);
  
  console.log(`👁️ Started monitoring directory: ${dirPath} (polling every ${POLL_INTERVAL/1000}s)`);
}
