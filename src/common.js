import fetch from 'node-fetch';
import UserAgent from 'user-agents';

export const CORS_HEADERS = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, OPTIONS',
  'Access-Control-Allow-Headers': '*',
};

const DATAGOKR_SERVICEKEY = process.env.DATAGOKR_SERVICEKEY;

/**
 * Creates an Express middleware that proxies GET requests to a specified base URL, injecting an API service key and handling CORS.
 *
 * The middleware only processes requests whose path is included in the provided set of allowed paths. It appends the required service key as a query parameter, sets a randomized User-Agent header, and forwards the request to the target service. The response is relayed back to the client with appropriate content type and CORS headers. If the service key environment variable is missing or the fetch fails, it responds with a 500 error and a JSON error message.
 *
 * @param {string} baseUrl - The base URL to which requests are proxied.
 * @param {Set<string>} allowedPaths - Set of request paths that are permitted to be proxied.
 * @returns {Function} An Express middleware function for proxying requests.
 */
export function createService(baseUrl, allowedPaths) {
  return async function(req, res, next) {
    if (!allowedPaths.has(req.path)) {
      return next();
    }

    if (!DATAGOKR_SERVICEKEY) {
      return res.status(500).json({
        error: 'Missing environment variable',
        message: 'DATAGOKR_SERVICEKEY is not defined in the environment'
      });
    }

    const params = new URLSearchParams(req.query);
    params.set('serviceKey', DATAGOKR_SERVICEKEY);
    const targetUrl = `${baseUrl}${req.path}?${params.toString()}`;

    const randomUserAgent = new UserAgent().toString();

    try {
      const response = await fetch(targetUrl, {
        method: 'GET',
        headers: { 'User-Agent': randomUserAgent },
        timeout: 2000,
      });

      res.set({
        'Content-Type': response.headers.get('content-type') || 'application/xml',
        ...CORS_HEADERS,
      });

      response.body.pipe(res);
    } catch (e) {
      console.error('Fetch Error:', e);
      res.status(500).json({
        error: 'Fetch Error',
        message: e.message,
      });
    }
  };
}
