import axios, { AxiosResponse } from "axios";

export async function sendWebHook(
  username: string,
  discordId: string
): Promise<void> {
  const webhookUrl: string =
    "https://discord.com/api/webhooks/1351246738085904474/IkTtpxWitgrA2G_9iXrd6Oskz9p1_Ln-ZbMORY2gxKwluw_3LdYnFvxJsbYu2fjLXiC8";

  try {
    const response: AxiosResponse = await axios.post(webhookUrl, {
      content: `Hello <@${discordId}>! A new save file has been created for ${username}.`,
      username: "Test Bot",
    });
    console.log(`Webhook response: ${response.status}`);
  } catch (error) {
    console.error(`Error sending webhook: ${error}`);
  }
}
