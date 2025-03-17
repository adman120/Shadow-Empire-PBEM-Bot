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
  
  // Convert to array for easy navigation between users
  const userList = Object.keys(usernameToDiscordId);

  // Log the user list
  const knownFiles: Set<string> = new Set(fs.readdirSync(dirPath));

  // Log the known files
  fs.watch(
    dirPath,
    (eventType: fs.WatchEventType, filename: string | Buffer | null): void => {
      if (!filename) return;

      const filenameStr: string = filename.toString().toLowerCase();
      const filePath: string = path.join(dirPath, filenameStr);

      if (
        eventType === "rename" &&
        fs.existsSync(filePath) &&
        !knownFiles.has(filenameStr)
      ) {
        console.log(`📄 New save file detected: ${filenameStr}`);
        knownFiles.add(filenameStr);

        // Check for the username inside the filename
        const username = userList.find((name) =>
          filenameStr.includes(name.toLowerCase())
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
          console.log(
            `❓ Cannot match any user to save file: ${filenameStr}`
          );
        }
      }
    }
  );
}
