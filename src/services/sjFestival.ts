import { createService } from "../common.js";

/**
 * Creates an Express middleware for proxying requests to the Sejong Festival API.
 * Access is restricted to a specific set of endpoints.
 */
export default function createSjFestival() {
  const baseUrl = "https://apis.data.go.kr/5690000/sjFestival";
  const allowedPaths = new Set(["/sj_00000360"]);
  return createService(baseUrl, allowedPaths);
}
