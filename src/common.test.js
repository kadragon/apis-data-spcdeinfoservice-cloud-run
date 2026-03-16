import { PassThrough } from "node:stream";
import { describe, expect, it, vi } from "vitest";

process.env.DATAGOKR_SERVICEKEY =
  process.env.DATAGOKR_SERVICEKEY || "test-service-key";

vi.mock("node-fetch", () => ({
  default: vi.fn(),
}));

vi.mock("user-agents", () => ({
  default: class {
    toString() {
      return "mock-user-agent";
    }
  },
}));

const { default: fetch } = await import("node-fetch");
const { createService } = await import("./common.js");

function createMockRes() {
  const res = new PassThrough();
  res.set = vi.fn();
  res.status = vi.fn().mockReturnValue(res);
  res.json = vi.fn();
  return res;
}

describe("createService", () => {
  it("proxies request for an allowed path with correct URL, headers, and streaming", async () => {
    const responseBody = new PassThrough();
    fetch.mockResolvedValueOnce({
      headers: {
        get: (name) => (name === "content-type" ? "application/xml" : null),
      },
      body: responseBody,
    });

    const baseUrl = "https://api.example.com";
    const allowedPaths = new Set(["/getAllowed"]);
    const middleware = createService(baseUrl, allowedPaths);

    const req = { path: "/getAllowed", query: { numOfRows: "10" } };
    const res = createMockRes();
    const next = vi.fn();

    await middleware(req, res, next);

    expect(next).not.toHaveBeenCalled();
    expect(fetch).toHaveBeenCalledOnce();

    const calledUrl = fetch.mock.calls[0][0];
    expect(calledUrl).toContain("https://api.example.com/getAllowed?");
    expect(calledUrl).toContain("numOfRows=10");
    expect(calledUrl).toContain("serviceKey=test-service-key");

    const calledOptions = fetch.mock.calls[0][1];
    expect(calledOptions.method).toBe("GET");
    expect(calledOptions.headers["User-Agent"]).toBe("mock-user-agent");

    expect(res.set).toHaveBeenCalledWith(
      expect.objectContaining({
        "Content-Type": "application/xml",
        "Access-Control-Allow-Origin": "*",
      }),
    );
  });

  it("calls next() and does not fetch for disallowed paths", async () => {
    fetch.mockClear();
    const middleware = createService(
      "https://api.example.com",
      new Set(["/allowed"]),
    );

    const req = { path: "/notAllowed", query: {} };
    const res = createMockRes();
    const next = vi.fn();

    await middleware(req, res, next);

    expect(next).toHaveBeenCalledOnce();
    expect(fetch).not.toHaveBeenCalled();
  });

  it("responds with 500 and timeout message on AbortError", async () => {
    const abortError = new DOMException("signal timed out", "AbortError");
    fetch.mockRejectedValueOnce(abortError);

    const middleware = createService(
      "https://api.example.com",
      new Set(["/path"]),
    );
    const req = { path: "/path", query: {} };
    const res = createMockRes();
    const next = vi.fn();

    await middleware(req, res, next);

    expect(res.status).toHaveBeenCalledWith(500);
    expect(res.json).toHaveBeenCalledWith({
      error: "Request Timeout",
      message: "Request timed out",
    });
  });

  it("responds with 500 and service unavailable on generic fetch error", async () => {
    fetch.mockRejectedValueOnce(new Error("network failure"));

    const middleware = createService(
      "https://api.example.com",
      new Set(["/path"]),
    );
    const req = { path: "/path", query: {} };
    const res = createMockRes();
    const next = vi.fn();

    await middleware(req, res, next);

    expect(res.status).toHaveBeenCalledWith(500);
    expect(res.json).toHaveBeenCalledWith({
      error: "Service Unavailable",
      message: "Unable to process request",
    });
  });
});
