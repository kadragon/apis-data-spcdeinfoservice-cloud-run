import { createService } from '../common.js';

export default function createSpcdeInfoService() {
  const baseUrl = 'https://apis.data.go.kr/B090041/openapi/service/SpcdeInfoService';
  const allowedPaths = new Set(['/getRestDeInfo', '/getAnniversaryInfo', '/get24DivisionsInfo']);
  return createService(baseUrl, allowedPaths);
}
