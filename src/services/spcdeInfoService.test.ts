import { PassThrough } from "node:stream";
import { describe, expect, it, vi } from "vitest";

process.env.DATAGOKR_SERVICEKEY = "test-key";

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

const { default: createSpcdeInfoService } = await import(
  "./spcdeInfoService.js"
);

const ALLOWED_PATHS = [
  "/getRestDeInfo",
  "/getAnniversaryInfo",
  "/get24DivisionsInfo",
];

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

describe("createSpcdeInfoService", () => {
  it("returns a middleware function", () => {
    const middleware = createSpcdeInfoService();
    expect(typeof middleware).toBe("function");
  });

  it("calls next for disallowed paths", async () => {
    const middleware = createSpcdeInfoService();
    const next = vi.fn();

    await middleware({ path: "/notAllowed", query: {} }, createMockRes(), next);
    expect(next).toHaveBeenCalledOnce();
  });

  it("does not call next for allowed paths", async () => {
    const middleware = createSpcdeInfoService();
    const next = vi.fn();

    for (const path of ALLOWED_PATHS) {
      next.mockClear();
      await middleware({ path, query: {} }, createMockRes(), next);
      expect(next).not.toHaveBeenCalled();
    }
  });
});
