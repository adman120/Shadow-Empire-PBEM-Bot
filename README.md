<div align="center">

# Shadow Empire PBEM Bot

_A Discord bot for automating player turns in Shadow Empire Play-By-Email (PBEM) games. Works in conjunction with file synchronization tools like Dropbox, Google Drive, or SyncThing._

</div>

<div align="center">

![GitHub Repo stars](https://img.shields.io/github/stars/1Solon/shadow-empire-pbem-bot?style=for-the-badge)
![GitHub forks](https://img.shields.io/github/forks/1Solon/shadow-empire-pbem-bot?style=for-the-badge)

</div>

## ✨ Features

- Monitors a directory for new Shadow Empire save files
- Automatically detects which player just completed their turn
- Notifies the next player via Discord webhook when it's their turn
- Automatically detects if a save file is misnamed and informs the player
- Configurable file name pattern matching and debouncing
- Runs in Docker for easy deployment
- Lightweight and efficient

## 🚀 Installation

### Docker (Recommended)

Pull the latest image from GitHub Container Registry:

```sh
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
      - USER_MAPPINGS=Player1 123456789012345678,Player2 234567890123456789
      - GAME_NAME=PBEM1
      - DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/your-webhook-url
      - WATCH_DIRECTORY=/app/data
      - IGNORE_PATTERNS=backup,temp
      - FILE_DEBOUNCE_MS=30000
    restart: unless-stopped
```

### Go Installation

To build and run locally:

```sh
git clone https://github.com/1Solon/shadow-empire-pbem-bot.git
cd shadow-empire-pbem-bot
go build -o shadow-empire-bot .
```

## 📚 Environment Variables

| Variable              | Description                                                                      | Required | Default  |
| --------------------- | -------------------------------------------------------------------------------- | -------- | -------- |
| `USER_MAPPINGS`       | Comma-separated list of usernames and Discord IDs (format: `Username DiscordID`) | Yes      | None     |
| `GAME_NAME`           | Name prefix for save files                                                       | No       | "pbem1"  |
| `DISCORD_WEBHOOK_URL` | Discord webhook URL for notifications                                            | Yes      | None     |
| `WATCH_DIRECTORY`     | Directory to monitor for save files                                              | No       | "./data" |
| `IGNORE_PATTERNS`     | Comma-separated patterns to ignore in filenames                                  | No       | None     |
| `FILE_DEBOUNCE_MS`    | Milliseconds to wait after file detection before processing                      | No       | 30000    |

### .env File Support

The bot also supports loading environment variables from a `.env` file. Create a file named `.env` in the same directory as the bot executable (or in your mounted `/app` directory when using Docker):

```
USER_MAPPINGS=Player1 123456789012345678,Player2 234567890123456789
GAME_NAME=PBEM1
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/your-webhook-url
WATCH_DIRECTORY=./data
IGNORE_PATTERNS=backup,temp
FILE_DEBOUNCE_MS=30000
```

## 📖 Usage

### Setup

1. Create a Discord webhook in your server
2. Configure environment variables with player names and their Discord IDs (either through environment variables or a .env file)
3. Set up a shared folder for save files using a file synchronization tool (Dropbox, Google Drive, SyncThing, etc.)
   - Recommendation: Use Shadow Empire's default save location: `C:\Users\<username>\Documents\My Games\Shadow Empire\<game name>`
4. Run the bot pointing to this shared folder

### Running with Docker

```sh
docker run -d \
  -e USER_MAPPINGS="Player1 123456789012345678,Player2 234567890123456789" \
  -e GAME_NAME="PBEM1" \
  -e DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/your-webhook-url" \
  -v "C:/Users/<username>/Documents/My Games/Shadow Empire/<game name>:/app/data" \
  ghcr.io/1solon/shadow-empire-pbem-bot:latest
```

### Running from source

```sh
export USER_MAPPINGS="Player1 123456789012345678,Player2 234567890123456789"
export GAME_NAME="PBEM1"
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/your-webhook-url"
export WATCH_DIRECTORY="C:/Users/<username>/Documents/My Games/Shadow Empire/<game name>"
./shadow-empire-bot
```

### Save File Naming Convention

The main Shadow Empire multiplayer community uses these naming formats:

```
PBEM1_turn1_Player1
```

or

```
PBEM1_Player1_turn1
```

The number in PBEM1 can be incremented for different game instances (PBEM2, PBEM3, etc.)
