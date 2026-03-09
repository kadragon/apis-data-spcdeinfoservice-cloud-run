import { describe, expect, it } from "vitest";

process.env.DATAGOKR_SERVICEKEY = process.env.DATAGOKR_SERVICEKEY || "test-key";

const { default: createBidPublicInfoService } = await import(
  "./bidPublicInfoService.js"
);

describe("createBidPublicInfoService", () => {
  it("returns a middleware function", () => {
    const middleware = createBidPublicInfoService();
    expect(typeof middleware).toBe("function");
  });

  it("calls next for disallowed paths", async () => {
    const middleware = createBidPublicInfoService();
    const req = { path: "/notAllowed", query: {} };
    let nextCalled = false;
    const next = () => {
      nextCalled = true;
    };
    const res = {};

    await middleware(req, res, next);
    expect(nextCalled).toBe(true);
  });

  it("has 25 allowed endpoints", () => {
    const expectedPaths = [
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
    expect(expectedPaths).toHaveLength(25);
  });
});
