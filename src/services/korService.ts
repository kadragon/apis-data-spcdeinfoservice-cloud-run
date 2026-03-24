import { createService } from "../common.js";

/**
 * Creates a service client for the KorService2 API (국문 관광정보 서비스) with access restricted to specific endpoints.
 * @return {object} A configured service client for the KorService2 API.
 */
export default function createKorService() {
  const baseUrl = "https://apis.data.go.kr/B551011/KorService2";
  const allowedPaths = new Set([
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
  ]);
  return createService(baseUrl, allowedPaths);
}
