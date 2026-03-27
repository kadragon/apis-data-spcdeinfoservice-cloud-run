import { Readable } from "node:stream";
import type { NextFunction, Request, Response } from "express";
import UserAgent from "user-agents";

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

class UpstreamError extends Error {
  constructor(
    message: string,
    public readonly statusCode: number,
    public readonly isTimeout: boolean,
  ) {
    super(message);
    this.name = "UpstreamError";
  }
}

const MAX_RETRIES = 2;
const TIMEOUT_MS = 10_000;
const BACKOFF_BASE_MS = 1_000;

async function fetchWithRetry(
  url: string,
  headers: Record<string, string>,
): Promise<globalThis.Response> {
  let lastError: unknown;

  for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
    try {
      const response = await fetch(url, {
        method: "GET",
        headers,
        signal: AbortSignal.timeout(TIMEOUT_MS),
      });

      if (response.status < 500) {
        return response;
      }

      await response.arrayBuffer();
      lastError = new UpstreamError(
        `Upstream returned ${response.status}`,
        502,
        false,
      );
    } catch (e) {
      const isTimeout = e instanceof Error && e.name === "AbortError";
      const message = e instanceof Error ? e.message : String(e);
      lastError = new UpstreamError(
        isTimeout ? "Request timed out" : message,
        isTimeout ? 504 : 502,
        isTimeout,
      );
    }

    if (attempt < MAX_RETRIES) {
      await sleep(BACKOFF_BASE_MS * 2 ** attempt);
    }
  }

  throw lastError;
}

export const CORS_HEADERS: Record<string, string> = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "GET, OPTIONS",
  "Access-Control-Allow-Headers": "x-api-key, Content-Type",
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
      const upstream = await fetchWithRetry(targetUrl, {
        "User-Agent": randomUserAgent,
      });

      res.status(upstream.status).set({
        "Content-Type":
          upstream.headers.get("content-type") || "application/xml",
        ...CORS_HEADERS,
      });

      if (!upstream.body) {
        res.end();
        return;
      }

      // @ts-expect-error: ReadableStream type mismatch between Web API and Node.js
      Readable.fromWeb(upstream.body)
        .pipe(res)
        .on("error", (err: Error) => {
          console.error("Pipe Error:", err);
          if (!res.headersSent) {
            res.status(500).end();
          } else {
            res.end();
          }
        });
    } catch (e) {
      console.error("Fetch Error:", e);
      res.set(CORS_HEADERS);
      if (e instanceof UpstreamError) {
        res.status(e.statusCode).json({
          error: e.isTimeout ? "Gateway Timeout" : "Bad Gateway",
          message: e.isTimeout
            ? "Request timed out"
            : "Unable to process request",
        });
      } else {
        res.status(502).json({
          error: "Bad Gateway",
          message: "Unable to process request",
        });
      }
    }
  };
}
