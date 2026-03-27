import { PassThrough } from "node:stream";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

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
    vi.useFakeTimers();
    fetchMock.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
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

  it("responds with 504 and timeout message after all retries exhausted on AbortError", async () => {
    const abortError = new DOMException("signal timed out", "AbortError");
    fetchMock.mockRejectedValue(abortError);

    const middleware = createService(
      "https://api.example.com",
      new Set(["/path"]),
    );
    const req = { path: "/path", query: {} };
    const res = createMockRes();
    const next = vi.fn();

    const promise = middleware(req, res, next);
    await vi.advanceTimersByTimeAsync(1000);
    await vi.advanceTimersByTimeAsync(2000);
    await promise;

    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(res.set).toHaveBeenCalledWith(
      expect.objectContaining({ "Access-Control-Allow-Origin": "*" }),
    );
    expect(res.status).toHaveBeenCalledWith(504);
    expect(res.json).toHaveBeenCalledWith({
      error: "Gateway Timeout",
      message: "Request timed out",
    });
  });

  it("responds with 502 and bad gateway on generic fetch error after all retries", async () => {
    fetchMock.mockRejectedValue(new Error("network failure"));

    const middleware = createService(
      "https://api.example.com",
      new Set(["/path"]),
    );
    const req = { path: "/path", query: {} };
    const res = createMockRes();
    const next = vi.fn();

    const promise = middleware(req, res, next);
    await vi.advanceTimersByTimeAsync(1000);
    await vi.advanceTimersByTimeAsync(2000);
    await promise;

    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(res.set).toHaveBeenCalledWith(
      expect.objectContaining({ "Access-Control-Allow-Origin": "*" }),
    );
    expect(res.status).toHaveBeenCalledWith(502);
    expect(res.json).toHaveBeenCalledWith({
      error: "Bad Gateway",
      message: "Unable to process request",
    });
  });

  it("retries on 5xx and succeeds on subsequent attempt", async () => {
    const responseBody = new PassThrough();
    fetchMock
      .mockResolvedValueOnce({
        status: 503,
        headers: { get: () => null },
        body: null,
        arrayBuffer: () => Promise.resolve(new ArrayBuffer(0)),
      })
      .mockResolvedValueOnce({
        status: 200,
        headers: {
          get: (name: string) =>
            name === "content-type" ? "application/xml" : null,
        },
        body: responseBody,
      });

    const middleware = createService(
      "https://api.example.com",
      new Set(["/path"]),
    );
    const req = { path: "/path", query: {} };
    const res = createMockRes();
    const next = vi.fn();

    const promise = middleware(req, res, next);
    await vi.advanceTimersByTimeAsync(1000);
    await promise;

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(res.status).toHaveBeenCalledWith(200);
  });

  it("does not retry on 4xx responses", async () => {
    fetchMock.mockResolvedValueOnce({
      status: 404,
      headers: {
        get: (name: string) =>
          name === "content-type" ? "application/json" : null,
      },
      body: null,
    });

    const middleware = createService(
      "https://api.example.com",
      new Set(["/path"]),
    );
    const req = { path: "/path", query: {} };
    const res = createMockRes();
    const next = vi.fn();

    await middleware(req, res, next);

    expect(fetchMock).toHaveBeenCalledOnce();
    expect(res.status).toHaveBeenCalledWith(404);
  });

  it("responds with 502 after all retries exhausted on 5xx", async () => {
    fetchMock.mockResolvedValue({
      status: 503,
      headers: { get: () => null },
      body: null,
      arrayBuffer: () => Promise.resolve(new ArrayBuffer(0)),
    });

    const middleware = createService(
      "https://api.example.com",
      new Set(["/path"]),
    );
    const req = { path: "/path", query: {} };
    const res = createMockRes();
    const next = vi.fn();

    const promise = middleware(req, res, next);
    await vi.advanceTimersByTimeAsync(1000);
    await vi.advanceTimersByTimeAsync(2000);
    await promise;

    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(res.status).toHaveBeenCalledWith(502);
    expect(res.json).toHaveBeenCalledWith({
      error: "Bad Gateway",
      message: "Unable to process request",
    });
  });
});
