import axios, { AxiosResponse } from 'axios';

async function sendWebhook(): Promise<void> {
  const webhookUrl: string = 'https://discord.com/api/webhooks/1351246738085904474/IkTtpxWitgrA2G_9iXrd6Oskz9p1_Ln-ZbMORY2gxKwluw_3LdYnFvxJsbYu2fjLXiC8';

  try {
    const response: AxiosResponse = await axios.post(webhookUrl, {
      content: 'Hello from Node.js using TypeScript!',
      username: 'MyBot',
    });
    console.log(`Status: ${response.status} - ${response.statusText}`);
  } catch (error) {
    console.error(error);
  }
}

sendWebhook();
