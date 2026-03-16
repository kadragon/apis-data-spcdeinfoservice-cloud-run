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

const { default: createBidPublicInfoService } = await import(
  "./bidPublicInfoService.js"
);

const ALLOWED_PATHS = [
  "/getBidPblancListInfoCnstwk",
  "/getBidPblancListInfoServc",
  "/getBidPblancListInfoFrgcpt",
  "/getBidPblancListInfoThng",
  "/getBidPblancListInfoThngBsisAmount",
  "/getBidPblancListInfoEtcPPSSrch",
  "/getBidPblancListPPIFnlRfpIssAtchFileInfo",
  "/getBidPblancListBidPrceCalclAInfo",
  "/getBidPblancListEvaluationIndstrytyMfrcInfo",
  "/getBidPblancListInfoEtc",
  "/getBidPblancListInfoEorderAtchFileInfo",
  "/getBidPblancListInfoFrgcptPurchsObjPrdct",
  "/getBidPblancListInfoServcPurchsObjPrdct",
  "/getBidPblancListInfoThngPurchsObjPrdct",
  "/getBidPblancListInfoPrtcptPsblRgn",
  "/getBidPblancListInfoLicenseLimit",
  "/getBidPblancListInfoCnstwkBsisAmount",
  "/getBidPblancListInfoThngPPSSrch",
  "/getBidPblancListInfoFrgcptPPSSrch",
  "/getBidPblancListInfoServcPPSSrch",
  "/getBidPblancListInfoCnstwkPPSSrch",
  "/getBidPblancListInfoChgHstryServc",
  "/getBidPblancListInfoChgHstryCnstwk",
  "/getBidPblancListInfoChgHstryThng",
  "/getBidPblancListInfoServcBsisAmount",
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

describe("createBidPublicInfoService", () => {
  it("returns a middleware function", () => {
    const middleware = createBidPublicInfoService();
    expect(typeof middleware).toBe("function");
  });

  it("calls next for disallowed paths", async () => {
    const middleware = createBidPublicInfoService();
    const req = { path: "/notAllowed", query: {} };
    const next = vi.fn();

    await middleware(req, createMockRes(), next);
    expect(next).toHaveBeenCalledOnce();
  });

  it("does not call next for allowed paths", async () => {
    const middleware = createBidPublicInfoService();
    const next = vi.fn();

    for (const path of ALLOWED_PATHS) {
      next.mockClear();
      const req = { path, query: {} };
      await middleware(req, createMockRes(), next);
      expect(next).not.toHaveBeenCalled();
    }
  });

  it("has exactly 25 allowed endpoints", async () => {
    const middleware = createBidPublicInfoService();
    const next = vi.fn();

    let acceptedCount = 0;
    for (const path of ALLOWED_PATHS) {
      next.mockClear();
      const req = { path, query: {} };
      await middleware(req, createMockRes(), next);
      if (!next.mock.calls.length) acceptedCount++;
    }

    expect(acceptedCount).toBe(25);
    expect(ALLOWED_PATHS).toHaveLength(25);
  });
});
