package types

// DiscordWebhook represents the complete structure for a Discord webhook request
type DiscordWebhook struct {
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Content   string  `json:"content"`
	Embeds    []Embed `json:"embeds"`
}

// Embed represents an embedded rich content section in a Discord message
type Embed struct {
	Color     int       `json:"color"`
	Thumbnail Thumbnail `json:"thumbnail"`
	Fields    []Field   `json:"fields"`
	Footer    Footer    `json:"footer"`
	Timestamp string    `json:"timestamp"`
}

// Thumbnail represents an image thumbnail in a Discord embed
type Thumbnail struct {
	URL string `json:"url"`
}

// Field represents a field with name-value pair in a Discord embed
type Field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Footer represents the footer section of a Discord embed
type Footer struct {
	Text string `json:"text"`
}
