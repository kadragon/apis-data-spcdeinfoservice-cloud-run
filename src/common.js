import fetch from 'node-fetch';
import UserAgent from 'user-agents';

export const CORS_HEADERS = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, OPTIONS',
  'Access-Control-Allow-Headers': '*',
};

const DATAGOKR_SERVICEKEY = process.env.DATAGOKR_SERVICEKEY;

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
