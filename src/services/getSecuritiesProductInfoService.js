import { createService } from '../common.js';

/**
 * Creates a service instance for accessing specific securities product information endpoints.
 * 
 * The service is configured to interact with the government "GetSecuritiesProductInfoService" API and restricts access to ETF, ETN, and ELW price information endpoints.
 * @returns {object} A service instance for querying allowed securities product information endpoints.
 */
export default function createSecuritiesProductInfoService() {
  const baseUrl = 'https://apis.data.go.kr/1160100/service/GetSecuritiesProductInfoService';
  const allowedPaths = new Set(['/getETFPriceInfo', '/getETNPriceInfo', '/getELWPriceInfo']);
  return createService(baseUrl, allowedPaths);
}
