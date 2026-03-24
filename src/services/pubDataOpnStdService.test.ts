import { PassThrough } from "node:stream";
import { describe, expect, it, vi } from "vitest";

process.env.DATAGOKR_SERVICEKEY = process.env.DATAGOKR_SERVICEKEY || "test-key";

globalThis.fetch = vi.fn(() =>
  Promise.resolve({
    status: 200,
    headers: { get: () => "application/xml" },
    body: new PassThrough(),
  }),
) as unknown as typeof fetch;

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
      return "test-agent";
    }
  },
}));

const { default: createPubDataOpnStdService } = await import(
  "./pubDataOpnStdService.js"
);

const ALLOWED_PATHS = [
  "/getDataSetOpnStdBidPblancInfo",
  "/getDataSetOpnStdScsbidInfo",
  "/getDataSetOpnStdCntrctInfo",
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

describe("createPubDataOpnStdService", () => {
  it("returns a middleware function", () => {
    const middleware = createPubDataOpnStdService();
    expect(typeof middleware).toBe("function");
  });

  it("calls next for disallowed paths", async () => {
    const middleware = createPubDataOpnStdService();
    const req = { path: "/notAllowed", query: {} };
    const next = vi.fn();

    await middleware(req, createMockRes(), next);
    expect(next).toHaveBeenCalledOnce();
  });

  it("does not call next for allowed paths", async () => {
    const middleware = createPubDataOpnStdService();
    const next = vi.fn();

    for (const path of ALLOWED_PATHS) {
      next.mockClear();
      const req = { path, query: {} };
      await middleware(req, createMockRes(), next);
      expect(next).not.toHaveBeenCalled();
    }
  });

  it("has exactly 3 allowed endpoints", async () => {
    const middleware = createPubDataOpnStdService();
    const next = vi.fn();

    let acceptedCount = 0;
    for (const path of ALLOWED_PATHS) {
      next.mockClear();
      const req = { path, query: {} };
      await middleware(req, createMockRes(), next);
      if (!next.mock.calls.length) acceptedCount++;
    }

    expect(acceptedCount).toBe(3);
    expect(ALLOWED_PATHS).toHaveLength(3);
  });
});
