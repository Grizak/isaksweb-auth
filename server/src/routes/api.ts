import express from "express";
import mongoose from "mongoose";

const router = express.Router();

router.get("/health", (req, res) => {
  res
    .status(200)
    .json({
      status: "OK",
      timestamp: Date.now(),
      database:
        mongoose.connection.readyState === 1 ? "Connected" : "Disconnected",
    });
});

export default router;
