import nodeloggerg from "nodeloggerg";

const logger = nodeloggerg({
  serverConfig: {
    startWebServer: process.env.NODE_ENV !== "production",
    serverPort: 9001,
    enableSearch: true,
    authEnabled: false,
    enableRealtime: true,
  },
  enableMetrics: false,
  compressOldLogs: true,
});

export default logger;
