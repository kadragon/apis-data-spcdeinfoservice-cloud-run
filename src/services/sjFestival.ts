import { createService } from "../common.js";

/**
 * Creates a service client for the Sejong Festival API with access restricted to specific endpoints.
 * @return {object} A configured service client for the sjFestival API.
 */
export default function createSjFestival() {
  const baseUrl = "https://apis.data.go.kr/5690000/sjFestival";
  const allowedPaths = new Set(["/sj_00000360"]);
  return createService(baseUrl, allowedPaths);
}
