import { createService } from "../common.js";

/**
 * Creates a service client for the PubDataOpnStdService API (나라장터 공공데이터개방표준서비스) with access restricted to specific endpoints.
 * @return {object} A configured service client for the PubDataOpnStdService API.
 */
export default function createPubDataOpnStdService() {
  const baseUrl = "https://apis.data.go.kr/1230000/ao/PubDataOpnStdService";
  const allowedPaths = new Set([
    "/getDataSetOpnStdBidPblancInfo",
    "/getDataSetOpnStdScsbidInfo",
    "/getDataSetOpnStdCntrctInfo",
  ]);
  return createService(baseUrl, allowedPaths);
}
