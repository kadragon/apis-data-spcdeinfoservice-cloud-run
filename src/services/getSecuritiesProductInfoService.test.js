import { PassThrough } from "node:stream";
import { describe, expect, it, vi } from "vitest";

process.env.DATAGOKR_SERVICEKEY = process.env.DATAGOKR_SERVICEKEY || "test-key";

vi.mock("node-fetch", () => ({
  default: vi.fn(() =>
    Promise.resolve({
      headers: { get: () => "application/xml" },
      body: new PassThrough(),
    }),
  ),
}));

vi.mock("user-agents", () => ({
  default: class {
    toString() {
      return "test-agent";
    }
  },
}));

const { default: createSecuritiesProductInfoService } = await import(
  "./getSecuritiesProductInfoService.js"
);

const ALLOWED_PATHS = [
  "/getETFPriceInfo",
  "/getETNPriceInfo",
  "/getELWPriceInfo",
];

function createMockRes() {
  const res = new PassThrough();
  res.set = vi.fn();
  res.status = vi.fn().mockReturnValue(res);
  res.json = vi.fn();
  return res;
}

describe("createSecuritiesProductInfoService", () => {
  it("returns a middleware function", () => {
    const middleware = createSecuritiesProductInfoService();
    expect(typeof middleware).toBe("function");
  });

  it("calls next for disallowed paths", async () => {
    const middleware = createSecuritiesProductInfoService();
    const next = vi.fn();

    await middleware({ path: "/notAllowed", query: {} }, createMockRes(), next);
    expect(next).toHaveBeenCalledOnce();
  });

  it("does not call next for allowed paths", async () => {
    const middleware = createSecuritiesProductInfoService();
    const next = vi.fn();

    for (const path of ALLOWED_PATHS) {
      next.mockClear();
      await middleware({ path, query: {} }, createMockRes(), next);
      expect(next).not.toHaveBeenCalled();
    }
  });

  it("has exactly 3 allowed endpoints", async () => {
    const middleware = createSecuritiesProductInfoService();
    const next = vi.fn();

    let acceptedCount = 0;
    for (const path of ALLOWED_PATHS) {
      next.mockClear();
      await middleware({ path, query: {} }, createMockRes(), next);
      if (!next.mock.calls.length) acceptedCount++;
    }

    expect(acceptedCount).toBe(3);
    expect(ALLOWED_PATHS).toHaveLength(3);
  });
});
