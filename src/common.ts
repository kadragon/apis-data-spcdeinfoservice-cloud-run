import type { NextFunction, Request, Response } from "express";
import fetch from "node-fetch";
import UserAgent from "user-agents";

export const CORS_HEADERS: Record<string, string> = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "GET, OPTIONS",
  "Access-Control-Allow-Headers": "*",
};

const DATAGOKR_SERVICEKEY = process.env.DATAGOKR_SERVICEKEY;
if (!DATAGOKR_SERVICEKEY) {
  console.error(
    "FATAL ERROR: DATAGOKR_SERVICEKEY is not defined in the environment.",
  );
  process.exit(1);
}

/**
 * Creates an Express middleware that proxies GET requests to a specified base URL, injecting an API service key and handling CORS.
 *
 * The middleware only processes requests whose path is included in the provided set of allowed paths. It appends the required service key as a query parameter, sets a randomized User-Agent header, and forwards the request to the target service. The response is relayed back to the client with appropriate content type and CORS headers. If the service key environment variable is missing or the fetch fails, it responds with a 500 error and a JSON error message.
 *
 * @param {string} baseUrl - The base URL to which requests are proxied.
 * @param {Set<string>} allowedPaths - Set of request paths that are permitted to be proxied.
 * @returns {Function} An Express middleware function for proxying requests.
 */
export function createService(
  baseUrl: string,
  allowedPaths: Set<string>,
): (req: Request, res: Response, next: NextFunction) => Promise<void> {
  return async (req: Request, res: Response, next: NextFunction) => {
    if (!allowedPaths.has(req.path)) {
      next();
      return;
    }

    const params = new URLSearchParams(req.query as Record<string, string>);
    params.set("serviceKey", DATAGOKR_SERVICEKEY as string);
    const targetUrl = `${baseUrl}${req.path}?${params.toString()}`;

    const randomUserAgent = new UserAgent().toString();

    try {
      const response = await fetch(targetUrl, {
        method: "GET",
        headers: { "User-Agent": randomUserAgent },
        signal: AbortSignal.timeout(10000),
      });

      res.set({
        "Content-Type":
          response.headers.get("content-type") || "application/xml",
        ...CORS_HEADERS,
      });

      const stream = response.body?.pipe(res);
      stream?.on("error", (err: Error) => {
        console.error("Pipe Error:", err);
        // End response on pipe error to avoid hanging
        if (!res.headersSent) {
          res.status(500).end();
        } else {
          res.end();
        }
      });
    } catch (e) {
      // Log detailed error internally
      console.error("Fetch Error:", e);
      const isTimeoutError = (e as Error).name === "AbortError";
      res.status(500).json({
        error: isTimeoutError ? "Request Timeout" : "Service Unavailable",
        message: isTimeoutError
          ? "Request timed out"
          : "Unable to process request",
      });
    }
  };
}
