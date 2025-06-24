import express from "express";
import { CORS_HEADERS } from "./common.js";
import createSpcdeInfoService from "./services/spcdeInfoService.js";
import createSecuritiesProductInfoService from "./services/getSecuritiesProductInfoService.js";
import crypto from "crypto";

const AUTH_API_KEY = process.env.AUTH_API_KEY;
if (!AUTH_API_KEY) {
  console.error("Error: Missing required environment variable AUTH_API_KEY");
  process.exit(1);
}

const app = express();

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

app.use("/SpcdeInfoService", createSpcdeInfoService());
app.use(
  "/GetSecuritiesProductInfoService",
  createSecuritiesProductInfoService()
);

app.use((req, res) => {
  res.status(404).send("Not Found");
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Proxy server running on port ${PORT}`);
});
