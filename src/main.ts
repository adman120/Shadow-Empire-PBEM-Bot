import { monitorDirectory } from "./functions/monitor";
import dotenv from "dotenv";
import * as fs from "fs";
import * as path from "path";

function main(): void {
  // Check if required environment variables exist
  if (!process.env.USER_MAPPINGS || !process.env.GAME_NAME) {
    // If not, try to load from .env file
    const envPath = path.resolve(process.cwd(), ".env");
    if (fs.existsSync(envPath)) {
      console.log("📝 Loading environment variables from .env file");
      dotenv.config();
    } else {
      console.warn(
        "⚠️ No .env file found and required environment variables not set"
      );
    }
  } else {
    console.log("🔧 Using environment variables from system");
  }

  // Check if specific environment variables are set after potential loading
  if (!process.env.USER_MAPPINGS) {
    console.warn("⚠️ USER_MAPPINGS environment variable is not set, exiting");
    process.exit(1);
  }
  if (!process.env.GAME_NAME) {
    console.log(
      "ℹ️ GAME_NAME environment variable is not set, using default: col"
    );
  }
  if (!process.env.DISCORD_WEBHOOK_URL) {
    console.warn("⚠️ DISCORD_WEBHOOK_URL environment variable is not set, webhook notifications will fail");
  }
  
  // Check if WATCH_DIRECTORY is set
  if (!process.env.WATCH_DIRECTORY) {
    console.warn("⚠️ WATCH_DIRECTORY environment variable is not set, using default: ./data");
  }

  // Check if IGNORE_PATTERNS is set
  if (process.env.IGNORE_PATTERNS) {
    console.log(`🔍 Will ignore files containing patterns: ${process.env.IGNORE_PATTERNS}`);
  }

  // Start monitoring the directory, default to "./data"
  const directoryToWatch: string = process.env.WATCH_DIRECTORY || "./data";
  console.log(`👀 Monitoring directory: ${directoryToWatch}`);
  monitorDirectory(directoryToWatch);
}

main();
