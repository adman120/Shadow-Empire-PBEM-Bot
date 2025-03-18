import * as fs from "fs";
import * as path from "path";

import { sendWebHook } from "./webhook";
import { userparse } from "./userParser";

export function monitorDirectory(dirPath: string): void {
  // Get username to Discord ID mappings from environment variable
  const usernameToDiscordId: Record<string, string> =
    userparse("USER_MAPPINGS");

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
