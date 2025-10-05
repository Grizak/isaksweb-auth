import { Worker } from "worker_threads";
import path from "path";

const createWorker = (fileName: string) => {
  return new Worker(path.resolve(`./src/workers/${fileName}.js`));
};

// Function to create a worker based on the environment
export const EmailWorker = createWorker("email");

export const shutDown = async () => {
  await EmailWorker.terminate();
};

export const sendEmail = (
  to: string,
  subject: string,
  messageText: string,
  messageHTML: string
): Promise<{ success: boolean; messageId?: string; error?: string }> => {
  return new Promise((resolve) => {
    EmailWorker.once("message", (response) => {
      resolve(response);
    });
    EmailWorker.postMessage({ to, subject, messageText, messageHTML });
  });
};
