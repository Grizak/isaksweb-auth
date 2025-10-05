import User from "@/models/User";
import { RegisterRequest, RouteData, RouteDataRequest } from "@/types";
import logger from "@/utils/logger";
import express from "express";
import mongoose from "mongoose";
import fs from "fs/promises";
import path from "path";

const ROUTE_DATA: RouteData = JSON.parse(
  await fs.readFile(
    path.join(process.cwd(), "config", "routeData.json"),
    "utf-8"
  )
);

const router = express.Router();

router.get("/health", (req, res) => {
  res.status(200).json({
    status: "OK",
    timestamp: Date.now(),
    database:
      mongoose.connection.readyState === 1 ? "Connected" : "Disconnected",
  });
});

router.post("/register", async (req: RegisterRequest, res) => {
  const { username, age, email, password } = req.body;

  const requiredFields = { username, age, email, password };
  const missingFields = Object.entries(requiredFields)
    .filter(([_, value]) => !value)
    .map(([key, _]) => key);

  if (missingFields.length > 0) {
    logger.error!(`Missing fields: ${missingFields.join(", ")}`);
    return res.status(400).json({
      error: "Missing required fields",
      missingFields,
    });
  }

  const newUser = new User({
    email,
    password,
    age,
    username,
  });

  await newUser.save();
});

router.get(
  "/validate/availability/user/:type/:value",
  async (req: express.Request<{ type: string; value: string }>, res) => {
    const { type, value } = req.params;

    if (!value || value.trim().length === 0) {
      return res.status(400).json({
        error: "Value cannot be empty",
        type,
      });
    }

    const validTypes: string[] = ["username", "email"];
    if (!validTypes.includes(type)) {
      return res.status(400).json({
        error: "Invalid type",
        type: type,
      });
    }

    const filter: Record<string, string> = {};

    filter[type] = value;

    const user = await User.findOne(filter);

    res.json({ available: !user });
  }
);

router.get("/routeData", (req: RouteDataRequest, res) => {
  res.json(ROUTE_DATA[req.query.pathname]);
});

export default router;
