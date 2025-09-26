import express from "express";
import logger from "@/utils/logger";
import "dotenv/config";
import mongoose from "mongoose";
import { shutDown as shutDownWorkers } from "./workers";

const app = express();
const PORT: number = Number(process.env.PORT || 3000);
const MONGO_URI: string | undefined = process.env.MONGO_URI;

if (!MONGO_URI) {
  await logger.error!(
    "process.env.MONGO_URI is not set, please set it and restart the server again",
    MONGO_URI
  );
  process.exit(1);
}

if (isNaN(PORT)) {
  await logger.error!(
    "process.env.PORT is set but isn't a number: ",
    process.env.PORT
  );
  process.exit(1);
}

// Connect to mongodb
mongoose
  .connect(MONGO_URI)
  .then(() => {
    logger.info!("Connected to mongodb");
  })
  .catch((error) => {
    logger.error!("Failed to connect to mongodb:", error);
  });
// End of mongoose connection

app.use(express.static("frontend"));
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use("/api", (await import("./routes/api")).default);

app.use((req, res) => {
  res.sendFile("frontend/index.html", { root: process.cwd() });
});

const server = app.listen(PORT, () => {
  logger.info!(`Server started on http://localhost:${PORT}`);
});

process.on("SIGINT", async () => {
  logger.warn!("Shutting down gracefully");
  mongoose.disconnect();
  server.close();
  await shutDownWorkers();
  process.exit(0);
});
