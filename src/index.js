import express from "express";
import fetch from "node-fetch";

const app = express();
const BASE_URL = "https://apis.data.go.kr/B090041/openapi/service/SpcdeInfoService";

const allowedPaths = new Set(["/getRestDeInfo", "/getAnniversaryInfo", "/get24DivisionsInfo"]);

app.use(async (req, res) => {
  const urlPath = req.path;
  if (!allowedPaths.has(urlPath)) {
    return res.status(404).send("Not Found");
  }

  const datagokr_serviceKey = process.env.DATAGOKR_SERVICEKEY;
  if (!datagokr_serviceKey) {
    return res.status(500).json({
      error: "Missing environment variable",
      message: "DATAGOKR_SERVICEKEY is not defined in the environment"
    });
  }

  const params = new URLSearchParams(req.query);
  params.set("serviceKey", datagokr_serviceKey);
  const targetUrl = `${BASE_URL}${urlPath}?${params.toString()}`;

  const userAgents = [
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.6167.140 Safari/537.36"
  ];
  const randomUserAgent = userAgents[Math.floor(Math.random() * userAgents.length)];

  try {
    const response = await fetch(targetUrl, {
      method: "GET",
      headers: {
        'User-Agent': randomUserAgent,
      },
      timeout: 2000,
    });

    res.set({
      'Content-Type': response.headers.get('content-type') || 'application/xml',
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, OPTIONS',
      'Access-Control-Allow-Headers': '*',
      'Host': 'www.ncs.go.kr',
      'X-Forwarded-For': '',
      'X-Forwarded-Host': '',
      'X-Forwarded-Proto': '',
    });

    response.body.pipe(res);
  } catch (e) {
    console.error("Fetch Error:", e);
    res.status(500).json({
      error: "Fetch Error",
      message: e.message
    });
  }
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Proxy server running on port ${PORT}`);
});