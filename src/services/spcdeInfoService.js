import { createService } from '../common.js';

/**
 * Creates a service client for the SpcdeInfoService API with access restricted to specific endpoints.
 * @return {object} A configured service client for the SpcdeInfoService API.
 */
export default function createSpcdeInfoService() {
  const baseUrl = 'https://apis.data.go.kr/B090041/openapi/service/SpcdeInfoService';
  const allowedPaths = new Set(['/getRestDeInfo', '/getAnniversaryInfo', '/get24DivisionsInfo']);
  return createService(baseUrl, allowedPaths);
}
