## 기존 서비스 및 common.js 테스트 추가

> 기존 서비스(spcdeInfoService, getSecuritiesProductInfoService)와 핵심 프록시 로직(common.js)에 대한 테스트 커버리지 확보.

- [x] common.js createService — 허용된 경로 요청 시 프록시 동작 확인
- [x] common.js createService — 허용되지 않은 경로 요청 시 next() 호출 확인
- [x] common.js createService — 타임아웃 시 적절한 에러 응답 확인
- [x] common.js createService — fetch 실패 시 에러 응답 확인
- [x] spcdeInfoService — 허용 경로 목록 및 서비스 생성 확인
- [x] getSecuritiesProductInfoService — 허용 경로 목록 및 서비스 생성 확인

## JS → TS 변환

> JavaScript 프로젝트를 TypeScript로 변환하여 타입 안전성을 확보한다. 기존 동작은 그대로 유지한다.

### 인프라 설정

- [ ] typescript, @types/express, @types/node devDependencies 추가 및 설치
- [ ] tsconfig.json 생성 (ESM, NodeNext, outDir: dist, strict)
- [ ] package.json 수정 — start 스크립트를 `node dist/index.js`로, lint-staged glob을 `*.{ts,json}`으로 변경
- [ ] vitest.config.js → vitest.config.ts 변환
- [ ] biome.json에 TypeScript 포함 확인 (기본 지원이므로 변경 불필요할 수 있음)

### 소스 파일 변환

- [ ] src/common.js → src/common.ts (타입 추가: Request, Response, NextFunction, 반환 타입 명시)
- [ ] src/services/spcdeInfoService.js → src/services/spcdeInfoService.ts
- [ ] src/services/getSecuritiesProductInfoService.js → src/services/getSecuritiesProductInfoService.ts
- [ ] src/services/bidPublicInfoService.js → src/services/bidPublicInfoService.ts
- [ ] src/index.js → src/index.ts (import 경로에서 .js 확장자 조정)

### 테스트 변환

- [ ] src/services/bidPublicInfoService.test.js → src/services/bidPublicInfoService.test.ts (import 경로, mock 타입 조정)

### 빌드 & 배포

- [ ] Dockerfile 수정 — 빌드 스테이지 추가 (npm run build 후 dist/ 만 복사)
- [ ] npm run build (tsc) 스크립트 추가 및 빌드 확인
- [ ] 전체 테스트 통과 확인
- [ ] 전체 lint 통과 확인
- [ ] 기존 .js 소스 파일 삭제
