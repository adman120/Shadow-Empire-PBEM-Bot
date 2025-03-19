package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Discord webhook request structure
type DiscordWebhook struct {
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Content   string  `json:"content"`
	Embeds    []Embed `json:"embeds"`
}

type Embed struct {
	Color     int       `json:"color"`
	Thumbnail Thumbnail `json:"thumbnail"`
	Fields    []Field   `json:"fields"`
	Footer    Footer    `json:"footer"`
	Timestamp string    `json:"timestamp"`
}

type Thumbnail struct {
	URL string `json:"url"`
}

type Field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Footer struct {
	Text string `json:"text"`
}

// SendWebHook sends a Discord webhook notification to the next player
func SendWebHook(username, discordID, nextUser string) error {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")

	if webhookURL == "" {
		fmt.Println("❌ DISCORD_WEBHOOK_URL environment variable is not set")
		return fmt.Errorf("webhook URL not set")
	}

	gameName := os.Getenv("GAME_NAME")
	if gameName == "" {
		gameName = "pbem1"
	}

	// Create webhook payload
	payload := DiscordWebhook{
		Username:  "Shadow Empire Assistant",
		AvatarURL: "https://raw.githubusercontent.com/auricom/home-ops/main/docs/src/assets/logo.png",
		Content:   fmt.Sprintf("🎲 It's your turn, <@%s>!", discordID),
		Embeds: []Embed{
			{
				Color: 0xFFA500,
				Thumbnail: Thumbnail{
					URL: "https://upload.wikimedia.org/wikipedia/en/4/4f/Shadow_Empire_cover.jpg",
				},
				Fields: []Field{
					{
						Name:  "📋 Save File Instructions",
						Value: fmt.Sprintf("After completing your turn, please save the file as:\n```\n%s_turn#_%s\n```\n(Replace # with the current turn number)", gameName, nextUser),
					},
				},
				Footer: Footer{
					Text: "Made with ❤️ by Solon",
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
	}

	// Marshal JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Send request
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send Discord notification: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("✅ Discord notification sent to %s (%s) with status: %d\n", username, discordID, resp.StatusCode)
	return nil
}
