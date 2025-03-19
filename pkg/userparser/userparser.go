package userparser

import (
	"fmt"
	"os"
	"strings"
)

// ParseUsers parses username to Discord ID mappings from a comma-separated environment variable
// Format: "Username1 DiscordId1,Username2 DiscordId2"
func ParseUsers(envVarName string) map[string]string {
	userRecords := make(map[string]string)
	envVar := os.Getenv(envVarName)

	if envVar == "" {
		fmt.Printf("Environment variable %s not found.\n", envVarName)
		return userRecords
	}

	pairs := strings.Split(envVar, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), " ", 2)
		if len(parts) == 2 {
			username := parts[0]
			discordId := parts[1]
			if username != "" && discordId != "" {
				userRecords[username] = discordId
			}
		}
	}

	return userRecords
}
