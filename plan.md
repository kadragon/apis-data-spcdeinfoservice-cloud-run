## 기존 서비스 및 common.js 테스트 추가

> 기존 서비스(spcdeInfoService, getSecuritiesProductInfoService)와 핵심 프록시 로직(common.js)에 대한 테스트 커버리지 확보.

- [ ] common.js createService — 허용된 경로 요청 시 프록시 동작 확인
- [ ] common.js createService — 허용되지 않은 경로 요청 시 next() 호출 확인
- [ ] common.js createService — 타임아웃 시 적절한 에러 응답 확인
- [ ] common.js createService — fetch 실패 시 에러 응답 확인
- [ ] spcdeInfoService — 허용 경로 목록 및 서비스 생성 확인
- [ ] getSecuritiesProductInfoService — 허용 경로 목록 및 서비스 생성 확인
