# 기여 가이드 (Contributing Guide)

이 문서는 AgiesX 프로젝트의 Git 워크플로우, 브랜치 전략, 커밋 규칙을 정의합니다.
원활한 협업을 위해 아래 규칙을 준수해 주세요.

---

## 목차

1. [브랜치 전략](#1-브랜치-전략)
2. [작업 흐름](#2-작업-흐름)
3. [커밋 메시지 규칙](#3-커밋-메시지-규칙)
4. [Pull Request 가이드](#4-pull-request-가이드)
5. [브랜치 보호 규칙](#5-브랜치-보호-규칙)
6. [라이선스 준수 가이드](#6-라이선스-준수-가이드)

---

## 1. 브랜치 전략

### 브랜치 구조

| 브랜치 | 역할 | 보호 수준 |
|--------|------|-----------|
| `main` | 최종 배포/안정 버전 | 🔒 보호됨 |
| `develop` | 개발 통합 브랜치 | 🔒 보호됨 |
| `feature/*` | 새 기능 개발 | 자유 |
| `fix/*` | 버그 수정 | 자유 |
| `docs/*` | 문서 작업 | 자유 |
| `refactor/*` | 코드 리팩토링 | 자유 |
| `chore/*` | 기타 작업 (빌드, 설정 등) | 자유 |

### 네이밍 규칙

브랜치 이름은 **영어 소문자, 숫자, 하이픈(`-`)**만 사용합니다.

```
feature/기능명
fix/버그명
docs/문서명
refactor/정리명
chore/기타작업
```

**예시:**
- `feature/ui-guide`
- `fix/login-error`
- `docs/readme-update`
- `refactor/auth-cleanup`
- `chore/update-dependencies`

---

## 2. 작업 흐름

### 기본 워크플로우

```
main ← develop ← feature/* (또는 fix/*, docs/* 등)
```

### 단계별 가이드

#### Step 1: 저장소 클론 (최초 1회)

```bash
git clone https://github.com/AISWopensource/AgiesX.git
cd AgiesX
```

#### Step 2: develop 브랜치에서 작업 브랜치 생성

```bash
git checkout develop
git pull origin develop
git checkout -b feature/ui-guide
```

#### Step 3: 작업 및 커밋

```bash
# 코드 작업 후
git add .
git commit -m "feat: 초보자용 UI 가이드 추가"
```

#### Step 4: 원격 저장소에 푸시

```bash
git push origin feature/ui-guide
```

#### Step 5: Pull Request 생성

GitHub에서 PR 생성:
- **Base:** `develop`
- **Compare:** `feature/ui-guide`

#### Step 6: 리뷰 및 머지

- 최소 1명의 리뷰어 승인 후 머지
- 머지 완료 후 feature 브랜치는 삭제해도 됩니다

### 최종 배포 흐름

```
develop → main (관리자가 수행)
```

---

## 3. 커밋 메시지 규칙

[Conventional Commits](https://www.conventionalcommits.org/) 양식을 따릅니다.

### 기본 형식

```
<type>(<scope>): <subject>

<body>  ← 선택사항
```

### 커밋 타입 (Type)

| 타입 | 설명 | 예시 |
|------|------|------|
| `feat` | 새로운 기능 추가 | `feat: 로그인 페이지 추가` |
| `fix` | 버그 수정 | `fix: 인증 토큰 만료 오류 수정` |
| `docs` | 문서 수정 | `docs: README 업데이트` |
| `style` | 코드 포맷팅 (기능 변경 없음) | `style: 코드 정렬` |
| `refactor` | 리팩토링 (기능 변경 없음) | `refactor: 함수 구조 개선` |
| `test` | 테스트 추가/수정 | `test: 로그인 테스트 추가` |
| `chore` | 빌드, 설정 등 기타 작업 | `chore: 의존성 업데이트` |

### 작성 규칙

- **제목(subject):** 50자 이내, 명령조로 작성
- **본문(body):** 선택사항, 상세 설명이 필요할 때만 사용
- **한글/영어** 모두 가능하나 팀 내 통일 권장

### 예시

```bash
feat(frontend): 다크모드 토글 버튼 추가

- 헤더에 다크모드 전환 버튼 배치
- localStorage에 사용자 설정 저장
```

```bash
fix(backend): JWT 토큰 갱신 로직 수정
```

```bash
docs: 기여 가이드 한글화
```

---

## 4. Pull Request 가이드

### PR 제목 형식

커밋 메시지와 동일한 형식을 사용합니다.

```
feat: 로그인 페이지 추가
fix: 인증 오류 수정
docs: README 업데이트
```

### PR 체크리스트

PR 생성 시 아래 항목을 확인해주세요:

- [ ] 코드가 정상 동작하는지 확인
- [ ] 관련 테스트가 있다면 통과하는지 확인
- [ ] 린터/포맷터 오류 없음
- [ ] 문서 업데이트 필요 시 반영

### 리뷰 프로세스

1. PR 생성 후 팀원에게 리뷰 요청
2. 최소 **1명** 이상의 승인(Approve) 필요
3. 리뷰어의 피드백 반영 후 재요청
4. 승인 후 머지

---

## 5. 브랜치 보호 규칙

### main 브랜치

- ✅ PR 필수 (직접 push 금지)
- ✅ 최소 1명 승인 필요
- ✅ Force push 금지
- ✅ 삭제 금지

### develop 브랜치

- ✅ PR 필수 (직접 push 금지)
- ✅ Force push 금지
- ✅ 삭제 금지

### feature/* 브랜치

- 보호 규칙 없음 (자유롭게 작업)

---

## 6. 라이선스 준수 가이드

### 개요

AgiesX는 **MIT 라이선스**로 배포되며, 모든 의존성은 MIT 호환 라이선스를 사용합니다.

### 의존성 추가 시 주의사항

새 의존성 추가 시 아래 라이선스 호환성을 확인하세요.

#### ✅ 허용 라이선스

- MIT
- Apache-2.0
- BSD-2-Clause, BSD-3-Clause
- ISC
- MPL-2.0 (수정 없이 사용 시)
- 0BSD (퍼블릭 도메인)

#### ❌ 비호환 라이선스

- GPL, LGPL, AGPL (특별한 예외 없이)
- CC-BY-SA (코드에는 사용 불가, 데이터는 가능)
- Proprietary/Commercial 라이선스

### PR 머지 전 체크리스트

1. **의존성 업데이트:**
   ```bash
   cd backend && go mod tidy
   cd ../frontend && npm install
   ```

2. **라이선스 리포트 생성:**
   ```bash
   ./scripts/generate-licenses.sh
   ```

   자세한 내용은 [licenses/README.md](licenses/README.md) 참조

3. **라이선스 스캔:**
   ```bash
   osv-scanner scan --experimental-licenses="MIT,Apache-2.0,BSD-2-Clause,BSD-3-Clause,ISC,MPL-2.0" backend
   osv-scanner scan --experimental-licenses="MIT,Apache-2.0,BSD-2-Clause,BSD-3-Clause,ISC,MPL-2.0" frontend
   ```

### 라이선스 검증 도구

| 도구 | 용도 |
|------|------|
| `osv-scanner` | 보안 및 라이선스 스캔 (권장) |
| `license-checker` | npm 라이선스 검증 |
| `go-licenses` | Go 모듈 라이선스 추출 |
| `go list` | Go 모듈 검사 |

### Docker 빌드 시

Docker 빌드 시 라이선스 리포트가 자동 생성됩니다:

- **Backend:** `/opt/pentagi/licenses/backend/`
  - `dependencies.txt` - Go 모듈 목록
  - `licenses.csv` - 상세 라이선스 정보

- **Frontend:** `/opt/pentagi/licenses/frontend/`
  - `dependencies.json` - npm 의존성 트리
  - `licenses.json` - 상세 라이선스 데이터
  - `licenses.csv` - 라이선스 요약

---

## 문의

질문이 있으시면 아래로 연락해주세요:

- **Email:** info@pentagi.com 또는 info@vxcontrol.com
- **GitHub Issues:** [Issues 페이지](https://github.com/AISWopensource/AgiesX/issues)
