<div align="center">

# Shadow Empire PBEM Bot

_A Discord bot for automating player turns in Shadow Empire Play-By-Email (PBEM) games._

</div>

<div align="center">

![GitHub Repo stars](https://img.shields.io/github/stars/1Solon/shadow-empire-pbem-bot?style=for-the-badge)
![GitHub forks](https://img.shields.io/github/forks/1Solon/shadow-empire-pbem-bot?style=for-the-badge)

</div>

## ✨ Features

- Monitors a directory for new Shadow Empire save files
- Automatically detects which player just completed their turn
- Notifies the next player via Discord webhook when it's their turn
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
      - ./data:/app/data
    environment:
      - USER_MAPPINGS=Player1 123456789012345678,Player2 234567890123456789
      - GAME_NAME=campaign1
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
| `GAME_NAME`           | Name prefix for save files                                                       | No       | "col"    |
| `DISCORD_WEBHOOK_URL` | Discord webhook URL for notifications                                            | Yes      | None     |
| `WATCH_DIRECTORY`     | Directory to monitor for save files                                              | No       | "./data" |
| `IGNORE_PATTERNS`     | Comma-separated patterns to ignore in filenames                                  | No       | None     |
| `FILE_DEBOUNCE_MS`    | Milliseconds to wait after file detection before processing                      | No       | 30000    |

## 📖 Usage

### Setup

1. Create a Discord webhook in your server
2. Configure environment variables with player names and their Discord IDs
3. Set up a shared folder for save files (could be Dropbox, Google Drive, etc.)
4. Run the bot pointing to this shared folder

### Running with Docker

```sh
docker run -d \
  -e USER_MAPPINGS="Player1 123456789012345678,Player2 234567890123456789" \
  -e GAME_NAME="campaign1" \
  -e DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/your-webhook-url" \
  -v /path/to/saves:/app/data \
  ghcr.io/1solon/shadow-empire-pbem-bot:latest
```

### Running from source

```sh
export USER_MAPPINGS="Player1 123456789012345678,Player2 234567890123456789"
export GAME_NAME="campaign1"
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/your-webhook-url"
export WATCH_DIRECTORY="./data"
./shadow-empire-bot
```

### Save File Naming Convention

Players should save their files using the following format:

```
campaign1_turn1_Player1
```
