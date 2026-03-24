import { PassThrough } from "node:stream";
import { describe, expect, it, vi } from "vitest";

process.env.DATAGOKR_SERVICEKEY = "test-key";

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

const { default: createKorService } = await import("./korService.js");

const ALLOWED_PATHS = [
  "/areaCode2",
  "/detailPetTour2",
  "/categoryCode2",
  "/areaBasedList2",
  "/locationBasedList2",
  "/searchKeyword2",
  "/searchFestival2",
  "/searchStay2",
  "/detailCommon2",
  "/detailIntro2",
  "/detailInfo2",
  "/detailImage2",
  "/lclsSystmCode2",
  "/areaBasedSyncList2",
  "/ldongCode2",
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

describe("createKorService", () => {
  it("returns a middleware function", () => {
    const middleware = createKorService();
    expect(typeof middleware).toBe("function");
  });

  it("calls next for disallowed paths", async () => {
    const middleware = createKorService();
    const next = vi.fn();

    await middleware({ path: "/notAllowed", query: {} }, createMockRes(), next);
    expect(next).toHaveBeenCalledOnce();
  });

  it("does not call next for allowed paths", async () => {
    const middleware = createKorService();
    const next = vi.fn();

    for (const path of ALLOWED_PATHS) {
      next.mockClear();
      await middleware({ path, query: {} }, createMockRes(), next);
      expect(next).not.toHaveBeenCalled();
    }
  });
});
