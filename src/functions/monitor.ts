import * as fs from "fs";
import * as path from "path";

import { sendWebHook } from "./webhook";

const usernameToDiscordId: Record<string, string> = {
  Solon: "448210429119037450",
  // add more username-to-ID mappings here
};

export function monitorDirectory(dirPath: string): void {
  const knownFiles: Set<string> = new Set(fs.readdirSync(dirPath));

  fs.watch(
    dirPath,
    (eventType: fs.WatchEventType, filename: string | Buffer | null): void => {
      if (!filename) return;

      const filenameStr: string = filename.toString();
      const filePath: string = path.join(dirPath, filenameStr);

      if (
        eventType === "rename" &&
        fs.existsSync(filePath) &&
        !knownFiles.has(filenameStr)
      ) {
        console.log(`File created: ${filenameStr}`);
        knownFiles.add(filenameStr);

        // Check for the username inside the filename no matter its position
        const username = Object.keys(usernameToDiscordId).find((name) =>
          filenameStr.includes(name)
        );
        if (username) {
          const discordId = usernameToDiscordId[username];
          sendWebHook(username, discordId);
        } else {
          console.log(
            `No Discord ID found for any username in filename: ${filenameStr}`
          );
        }
      }
    }
  );
}
