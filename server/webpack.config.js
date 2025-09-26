import path from "path";
import { fileURLToPath } from "url";
import TerserPlugin from "terser-webpack-plugin";
import CopyPlugin from "copy-webpack-plugin";
import fs from "fs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const WORKERS_DIR = path.resolve(__dirname, "src", "workers");
const WORKERS = {};

if (fs.existsSync(WORKERS_DIR)) {
  const workerFiles = fs
    .readdirSync(WORKERS_DIR)
    .filter(
      (file) =>
        !file.endsWith("index.ts") &&
        (file.endsWith(".ts") || file.endsWith(".js"))
    );

  for (const file of workerFiles) {
    const workerName = path.basename(file, path.extname(file));
    WORKERS[`workers/${workerName}`] = `./src/workers/${file}`;
  }
}

export default {
  entry: {
    bundle: "./src/index.ts",
    ...WORKERS,
  },
  target: "node",
  module: {
    rules: [
      {
        test: /\.ts$/,
        use: "ts-loader",
        exclude: /node_modules/,
      },
    ],
  },
  resolve: {
    extensions: [".ts", ".js"],
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  output: {
    filename: "[name].cjs",
    path: path.resolve(__dirname, "dist"),
    clean: true,
  },
  plugins: [
    new CopyPlugin({
      patterns: [
        { from: "static", to: "static" },
        { from: "frontend", to: "frontend" },
      ],
    }),
  ],
  optimization: {
    minimize: true,
    minimizer: [new TerserPlugin()],
  },
  mode: "production",
};
