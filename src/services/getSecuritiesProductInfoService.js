import { createService } from '../common.js';

export default function createSecuritiesProductInfoService() {
  const baseUrl = 'https://apis.data.go.kr/1160100/service/GetSecuritiesProductInfoService';
  const allowedPaths = new Set(['/getETFPriceInfo', '/getETNPriceInfo', '/getELWPriceInfo']);
  return createService(baseUrl, allowedPaths);
}
