/**
 * Parse username to Discord ID mappings from a comma-separated environment variable
 * Format: "Username1 DiscordId1,Username2 DiscordId2"
 */
export function userparse(envVarName: string): Record<string, string> {
  const userRecords: Record<string, string> = {};
  const envVar = process.env[envVarName];
  
  if (!envVar) {
    console.log(`Environment variable ${envVarName} not found.`);
    return userRecords;
  }

  const pairs = envVar.split(',');
  for (const pair of pairs) {
    const [username, discordId] = pair.trim().split(' ');
    if (username && discordId) {
      userRecords[username] = discordId;
    }
  }
  
  return userRecords;
}
