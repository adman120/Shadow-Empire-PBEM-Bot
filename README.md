<div align="center">

# Shadow Empire PBEM Bot

_A Discord bot for automating player turns in Shadow Empire Play-By-Email (PBEM) games. Works in conjunction with file synchronization tools like Dropbox, Google Drive, or SyncThing._

![GitHub Repo stars](https://img.shields.io/github/stars/1Solon/shadow-empire-pbem-bot?style=for-the-badge)
![GitHub forks](https://img.shields.io/github/forks/1Solon/shadow-empire-pbem-bot?style=for-the-badge)

</div>

---

## ✨ Features

- Monitors a directory for new Shadow Empire save files
- Automatically detects which player just completed their turn
- Determines the current turn number
- Notifies the next player via Discord webhook when it's their turn
- Automatically detects if a save file is misnamed and informs the player
- Configurable file name pattern matching and debouncing
- Runs in Docker for easy deployment
- Lightweight and efficient

---

## 🚀 Installation

### Docker (Recommended)

Pull the latest image from GitHub Container Registry:

```bash
docker pull ghcr.io/1solon/shadow-empire-pbem-bot:latest
```

Or use Docker Compose:

```yaml
version: "3"
services:
  shadow-empire-bot:
    image: ghcr.io/1solon/shadow-empire-pbem-bot:latest
    volumes:
      # Map to Shadow Empire's default save location
      - "C:/Users/<username>/Documents/My Games/Shadow Empire/<game name>:/app/data"
    environment:
      - USER_MAPPINGS=1 Player1 123456789012345678,2 Player2 234567890123456789
      - GAME_NAME=PBEM1
      - DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/your-webhook-url
      - WATCH_DIRECTORY=/app/data
      - IGNORE_PATTERNS=backup,temp
      - FILE_DEBOUNCE_MS=30000
      - FILE_AGE_LIMIT=
      - FILE_CHECK_TIME=1
    restart: unless-stopped
```

### Go Installation

To build and run locally:

```bash
git clone https://github.com/1Solon/shadow-empire-pbem-bot.git
cd shadow-empire-pbem-bot
go build -o shadow-empire-bot .
```

---

## 📚 Environment Variables

| Variable              | Description                                                                                 | Required | Default  |
| :-------------------- | :------------------------------------------------------------------------------------------ | :------: | :------- |
| `USER_MAPPINGS`       | Comma-separated list of usernames and Discord IDs (format: `TurnNumber Username DiscordID`) |    ✅    | None     |
| `GAME_NAME`           | Name prefix for save files                                                                  |    ❌    | "pbem1"  |
| `DISCORD_WEBHOOK_URL` | Discord webhook URL for notifications                                                       |    ✅    | None     |
| `WATCH_DIRECTORY`     | Directory to monitor for save files                                                         |    ❌    | "./data" |
| `IGNORE_PATTERNS`     | Comma-separated patterns to ignore in filenames                                             |    ❌    | None     |
| `FILE_DEBOUNCE_MS`    | Milliseconds to wait after file detection before processing                                 |    ❌    | 30000    |

### .env File Support

The bot also supports loading environment variables from a `.env` file. Create a file named `.env` in the same directory as the bot executable (or in your mounted `/app` directory when using Docker):

```ini
USER_MAPPINGS=1 Player1 123456789012345678,2 Player2 234567890123456789
GAME_NAME=PBEM1
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/your-webhook-url
WATCH_DIRECTORY=./data
IGNORE_PATTERNS=backup,temp
FILE_DEBOUNCE_MS=30000
```

---

## 📖 Usage

### Setup

1. Create a Discord webhook in your server
2. Configure environment variables with player names and their Discord IDs
3. Set up a shared folder for save files using a file synchronization tool
4. Run the bot pointing to this shared folder

> **Recommendation:** Use Shadow Empire's default save location:  
> `C:\Users\<username>\Documents\My Games\Shadow Empire\<game name>`

### Understanding USER_MAPPINGS

The `USER_MAPPINGS` environment variable connects in-game player names with Discord user IDs:

```ini
USER_MAPPINGS=1 Player1 123456789012345678,2 Player2 234567890123456789
```

Each mapping follows the format `PlayerName DiscordUserID`, with multiple mappings separated by commas.

#### How to Get Discord User IDs

To get a Discord user ID:

1. Open Discord settings by clicking the gear icon near your username
2. Go to "Advanced" and enable "Developer Mode"
3. Right-click on the username of any user in your server
4. Select "Copy ID" from the context menu

The copied ID is a long number (e.g., 123456789012345678) that uniquely identifies that Discord user.

> **Important:** Make sure the in-game player names exactly match the names of the players in your Shadow Empire game.

### Running with Docker

```bash
docker run -d \
  -e USER_MAPPINGS="1 Player1 123456789012345678,2 Player2 234567890123456789" \
  -e GAME_NAME="PBEM1" \
  -e DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/your-webhook-url" \
  -v "C:/Users/<username>/Documents/My Games/Shadow Empire/<game name>:/app/data" \
  ghcr.io/1solon/shadow-empire-pbem-bot:latest
```

### Running from Source

```bash
export USER_MAPPINGS="1 Player1 123456789012345678,2 Player2 234567890123456789"
export GAME_NAME="PBEM1"
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/your-webhook-url"
export WATCH_DIRECTORY="C:/Users/<username>/Documents/My Games/Shadow Empire/<game name>"
./shadow-empire-bot
```

---

### Save File Naming Convention

The main Shadow Empire multiplayer community uses these naming formats:

```
PBEM1_turn1_Player1
```

or

```
PBEM1_Player1_turn1
```

> **Note:** The number in PBEM1 can be incremented for different game instances (PBEM2, PBEM3, etc.)
