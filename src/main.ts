// import axios, { AxiosResponse } from 'axios';
import { monitorDirectory } from "./functions/monitor";

function main(): void {
  const directoryToWatch: string = "./data";
  console.log(`Monitoring directory: ${directoryToWatch}`);
  monitorDirectory(directoryToWatch);
}

main();
