import crypto from "node:crypto";
import express from "express";
import { CORS_HEADERS } from "./common.js";
import createBidPublicInfoService from "./services/bidPublicInfoService.js";
import createSecuritiesProductInfoService from "./services/getSecuritiesProductInfoService.js";
import createKorService from "./services/korService.js";
import createPubDataOpnStdService from "./services/pubDataOpnStdService.js";
import createSpcdeInfoService from "./services/spcdeInfoService.js";

const AUTH_API_KEY = process.env.AUTH_API_KEY;
if (!AUTH_API_KEY) {
  console.error("Error: Missing required environment variable AUTH_API_KEY");
  process.exit(1);
}

const app = express();

app.get("/health", (_req, res) => {
  res.status(200).json({ status: "ok" });
});

app.use((req, res, next) => {
  if (req.method === "OPTIONS") {
    return res.set(CORS_HEADERS).status(204).send();
  }

  const clientApiKey = req.header("x-api-key");
  // Constant-time API key comparison to mitigate timing attacks
  const clientKey = Buffer.from(clientApiKey || "", "utf8");
  const authKey = Buffer.from(AUTH_API_KEY, "utf8");
  if (
    !clientApiKey ||
    clientKey.length !== authKey.length ||
    !crypto.timingSafeEqual(clientKey, authKey)
  ) {
    return res.status(401).json({
      error: "Unauthorized",
      message: "Invalid or missing x-api-key header",
    });
  }

  next();
});

app.use("/BidPublicInfoService", createBidPublicInfoService());
app.use("/KorService2", createKorService());
app.use("/PubDataOpnStdService", createPubDataOpnStdService());
app.use("/SpcdeInfoService", createSpcdeInfoService());
app.use(
  "/GetSecuritiesProductInfoService",
  createSecuritiesProductInfoService(),
);

app.use((_req, res) => {
  res.status(404).json({ error: "Not Found" });
});

const PORT = process.env.PORT || 3000;
const server = app.listen(PORT, () => {
  console.log(`Proxy server running on port ${PORT}`);
});

function shutdown() {
  console.log("Shutting down gracefully...");
  server.close((err) => {
    if (err) {
      console.error("Error shutting down server:", err);
      process.exit(1);
    }
    process.exit(0);
  });
  setTimeout(() => {
    console.error("Forced shutdown after timeout");
    process.exit(1);
  }, 5000).unref();
}

process.on("SIGTERM", shutdown);
process.on("SIGINT", shutdown);
