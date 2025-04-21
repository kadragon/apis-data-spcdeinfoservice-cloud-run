import express from "express";
import fetch from "node-fetch";
import UserAgent from "user-agents";

// 재사용 가능한 CORS 헤더 설정
const CORS_HEADERS = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, OPTIONS',
  'Access-Control-Allow-Headers': '*',
};

const AUTH_API_KEY = process.env.AUTH_API_KEY;
const DATAGOKR_SERVICEKEY = process.env.DATAGOKR_SERVICEKEY;

const app = express();
const BASE_URL = "https://apis.data.go.kr/B090041/openapi/service/SpcdeInfoService";

const allowedPaths = new Set(["/getRestDeInfo", "/getAnniversaryInfo", "/get24DivisionsInfo"]);

app.use(async (req, res) => {
  // CORS preflight handling
  if (req.method === 'OPTIONS') {
    return res.set(CORS_HEADERS).status(204).send();
  }

  // 인증용 API 키 검증
  const clientApiKey = req.header('x-api-key');
  if (!clientApiKey || clientApiKey !== AUTH_API_KEY) {
    return res.status(401).json({
      error: "Unauthorized",
      message: "Invalid or missing x-api-key header"
    });
  }

  // 공공데이터 포털 서비스 키 존재 확인 
  if (!DATAGOKR_SERVICEKEY) {
    return res.status(500).json({
      error: "Missing environment variable",
      message: "DATAGOKR_SERVICEKEY is not defined in the environment"
    });
  }
  
  // 요청 가능 주소 확인
  const urlPath = req.path;
  if (!allowedPaths.has(urlPath)) {
    return res.status(404).send("Not Found");
  }

  // 쿼리에 서비스 키 추가
  const params = new URLSearchParams(req.query);
  params.set("serviceKey", DATAGOKR_SERVICEKEY);
  const targetUrl = `${BASE_URL}${urlPath}?${params.toString()}`;

  // 임의 userAgent 생성
  const randomUserAgent = new UserAgent().toString();

  try {
    const response = await fetch(targetUrl, {
      method: "GET",
      headers: {
        'User-Agent': randomUserAgent
      },
      timeout: 2000,
    });

    res.set({
      'Content-Type': response.headers.get('content-type') || 'application/xml',
      ...CORS_HEADERS,
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