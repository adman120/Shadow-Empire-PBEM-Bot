import axios, { AxiosResponse } from "axios";

export async function sendWebHook(
  username: string,
  discordId: string,
  nextUser: string
): Promise<void> {
  const webhookUrl: string =
    "https://discord.com/api/webhooks/1351246738085904474/IkTtpxWitgrA2G_9iXrd6Oskz9p1_Ln-ZbMORY2gxKwluw_3LdYnFvxJsbYu2fjLXiC8";
  
  const gameName: string = process.env.GAME_NAME || "col";

  try {
    const response: AxiosResponse = await axios.post(
      webhookUrl,
      {
        username: "Shadow Empire Assistant",
        avatar_url: "https://raw.githubusercontent.com/auricom/home-ops/main/docs/src/assets/logo.png",
        content: `🎲 It's your turn, <@${discordId}>!`,
        embeds: [
          {
            color: 0xFFA500,
            thumbnail: {
              url: "https://upload.wikimedia.org/wikipedia/en/4/4f/Shadow_Empire_cover.jpg"
            },
            fields: [
              {
                name: "📋 Save File Instructions",
                value: `After completing your turn, please save the file as:\n\`\`\`\n${gameName}_turn#_${nextUser}\n\`\`\`\n(Replace # with the current turn number)`
              }
            ],
            footer: {
              text: "Made with ❤️ by Solon",
            },
            timestamp: new Date().toISOString()
          }
        ]
      },
      {
        headers: {
          "Content-Type": "application/json",
        },
      }
    );
    console.log(`✅ Discord notification sent to ${username} (${discordId}) with status: ${response.status}`);
  } catch (error) {
    console.error(`❌ Failed to send Discord notification: ${error}`);
  }
}
