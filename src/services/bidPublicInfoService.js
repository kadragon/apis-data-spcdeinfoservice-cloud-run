import { createService } from "../common.js";

/**
 * Creates a service client for the BidPublicInfoService API with access restricted to specific endpoints.
 * @return {object} A configured service client for the BidPublicInfoService API.
 */
export default function createBidPublicInfoService() {
  const baseUrl = "https://apis.data.go.kr/1230000/ad/BidPublicInfoService";
  const allowedPaths = new Set([
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
  ]);
  return createService(baseUrl, allowedPaths);
}
