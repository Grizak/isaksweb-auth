import { parentPort as exportParentPort } from "worker_threads";
import { sendEmail } from "./utils/mailer.js";

if (!exportParentPort) {
  console.error(
    "This file needs to be ran as a node.js worker, please do not invoke directly"
  );
  process.exit(1);
}

const parentPort = exportParentPort;

parentPort.on("message", async (message) => {
  const { to, subject, messageText, messageHTML } = message;
  try {
    const response = await sendEmail({
      to,
      subject,
      text: messageText,
      html: messageHTML,
    });
    parentPort.postMessage({
      success: response.success,
      messageId: response.messageId,
    });
  } catch (error) {
    parentPort.postMessage({
      success: false,
      error: error instanceof Error ? error.message : "Unknown error occurred",
    });
  }
});
