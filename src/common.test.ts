import { PassThrough } from "node:stream";
import { beforeEach, describe, expect, it, vi } from "vitest";

process.env.DATAGOKR_SERVICEKEY = "test-service-key";

const fetchMock = vi.fn();
globalThis.fetch = fetchMock;

vi.mock("node:stream", async (importOriginal) => {
  const actual = (await importOriginal()) as typeof import("node:stream");
  return {
    ...actual,
    Readable: {
      ...actual.Readable,
      fromWeb: (stream: unknown) => stream,
    },
  };
});

vi.mock("user-agents", () => ({
  default: class {
    toString() {
      return "mock-user-agent";
    }
  },
}));

const { createService } = await import("./common.js");

function createMockRes() {
  const res = new PassThrough() as PassThrough & {
    set: ReturnType<typeof vi.fn>;
    status: ReturnType<typeof vi.fn>;
    json: ReturnType<typeof vi.fn>;
  };
  res.set = vi.fn();
  res.status = vi.fn().mockReturnValue(res);
  res.json = vi.fn();
  return res;
}

describe("createService", () => {
  beforeEach(() => {
    fetchMock.mockReset();
  });

  it("proxies request for an allowed path with correct URL, headers, and streaming", async () => {
    const responseBody = new PassThrough();
    fetchMock.mockResolvedValueOnce({
      status: 200,
      headers: {
        get: (name: string) =>
          name === "content-type" ? "application/xml" : null,
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
    expect(fetchMock).toHaveBeenCalledOnce();

    const calledUrl = fetchMock.mock.calls[0][0];
    expect(calledUrl).toContain("https://api.example.com/getAllowed?");
    expect(calledUrl).toContain("numOfRows=10");
    expect(calledUrl).toContain("serviceKey=test-service-key");

    const calledOptions = fetchMock.mock.calls[0][1];
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
    const middleware = createService(
      "https://api.example.com",
      new Set(["/allowed"]),
    );

    const req = { path: "/notAllowed", query: {} };
    const res = createMockRes();
    const next = vi.fn();

    await middleware(req, res, next);

    expect(next).toHaveBeenCalledOnce();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("responds with 500 and timeout message on AbortError", async () => {
    const abortError = new DOMException("signal timed out", "AbortError");
    fetchMock.mockRejectedValueOnce(abortError);

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
    fetchMock.mockRejectedValueOnce(new Error("network failure"));

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
