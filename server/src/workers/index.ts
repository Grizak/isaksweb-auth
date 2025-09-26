import { Worker } from "worker_threads";
import path from "path";

const isDevelopment = process.env.NODE_ENV !== "production";

const createWorker = (fileName: string) => {
  if (isDevelopment) {
    // Development: use tsx to run TypeScript directly
    return new Worker("tsx", {
      argv: [path.resolve(`./src/workers/${fileName}.ts`)],
      execArgv: [],
    });
  } else {
    // Production: use compiled JavaScript
    return new Worker(path.resolve(`./dist/workers/${fileName}.cjs`));
  }
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
