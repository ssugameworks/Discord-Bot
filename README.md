# 알고리즘 행사 (가제)를 위한 디스코드 봇

백준(Baekjoon) 알고리즘 문제 풀이 점수를 집계하여 스코어보드를 제공하는 디스코드 봇입니다.

## 주요 기능

- 🎯 백준 사용자 자동 등록 및 티어 확인
- 📊 solved.ac API를 활용한 실시간 점수 계산  
- 🏆 자동 스코어보드 공개 (스케줄링 가능)
- 🔒 블랙아웃 모드 지원 (스코어보드 비공개)
- ⚡ 도전/기본/연습 문제에 따른 차등 점수 (1.4배/1.0배/0.5배)
- 🛠️ 대회 생성 및 관리 기능
- 💬 DM 및 서버 채널 모두 지원

## 점수 계산 방식

### 점수 공식
```
최종점수 = Σ(문제 난이도 점수 × 가중치)
```

### 가중치 적용
- **도전 문제** (현재 티어보다 높은 문제): 1.4배
- **기본 문제** (현재 티어와 같은 문제): 1.0배  
- **연습 문제** (현재 티어보다 낮은 문제): 0.5배

### 난이도별 점수표
| 티어 | 점수 | 티어 | 점수 | 티어 | 점수 |
|------|------|------|------|------|------|
| Bronze V-I | 1-5점 | Silver V-I | 8-16점 | Gold V-I | 18-25점 |
| Platinum V-I | 28-37점 | Diamond V-I | 40-50점 | Ruby V-I | 55-75점 |

## 설치 및 실행

### 1. 환경 설정
```bash
# Discord Bot Token 설정 (필수)
export DISCORD_BOT_TOKEN="your_discord_bot_token_here"

# Discord Channel ID 설정 (자동 스코어보드 전송용, 선택사항)  
export DISCORD_CHANNEL_ID="your_channel_id_here"
```

### 2. 의존성 설치
```bash
go mod tidy
```

### 3. 봇 실행
```bash
go run main.go
```

## Discord Bot 설정

### Bot 권한 설정
Discord Developer Portal에서 봇 생성 시 다음 권한이 필요합니다:
- `Send Messages` - 메시지 전송
- `Read Message History` - 메시지 기록 읽기
- `View Channels` - 채널 보기

### 인텐트 설정
다음 인텐트들이 활성화되어야 합니다:
- `Message Content Intent` - 메시지 내용 읽기
- `Server Members Intent` - 서버 멤버 정보 (관리자 권한 확인용)

## 사용법

### 참가자 명령어
- `!참가 <이름> <백준ID>` 또는 `!register <이름> <백준ID>` - 대회 참가 신청
- `!스코어보드` 또는 `!scoreboard` - 현재 스코어보드 확인 (서버에서만)
- `!참가자` 또는 `!participants` - 참가자 목록 확인
- `!도움말` 또는 `!help` - 도움말 표시
- `!ping` - 봇 응답 확인

### 관리자 명령어 (서버 관리자만)
- `!대회 create <대회명> <시작일> <종료일>` - 대회 생성
  - 예시: `!대회 create 2024알고리즘대회 2024-01-01 2024-01-21`
- `!대회 status` - 대회 상태 확인
- `!대회 blackout <on/off>` - 스코어보드 공개/비공개 설정
- `!대회 update <필드> <값>` - 대회 정보 수정
  - 필드: name, start, end
  - 예시: `!대회 update name 새로운대회명`

## 자동 스코어보드

- **전송 시간**: 매일 오전 9시 (환경변수로 설정 가능)
- **블랙아웃**: 대회 종료 3일 전부터 자동 비공개 또는 수동 설정
- **채널 설정**: `DISCORD_CHANNEL_ID` 환경변수로 지정
- **활성화 조건**: `DISCORD_CHANNEL_ID`가 설정된 경우에만 활성화

## 데이터 저장

봇은 JSON 파일을 사용하여 데이터를 저장합니다:
- `participants.json` - 참가자 정보
- `competition.json` - 대회 설정

## API 사용

### solved.ac API
- **사용자 정보**: `https://solved.ac/api/v3/user/show?handle={백준ID}`
- **TOP 100**: `https://solved.ac/api/v3/user/top_100?handle={백준ID}`

## 프로젝트 구조

```
discord-bot/
├── main.go              # 애플리케이션 진입점
├── app/
│   └── app.go          # 애플리케이션 생명주기 관리
├── config/
│   └── config.go        # 구조화된 환경 설정 관리
├── constants/
│   └── constants.go     # 전역 상수 정의
├── utils/
│   ├── logger.go        # 로깅 시스템
│   └── validation.go    # 유효성 검사 유틸리티
├── models/
│   └── participant.go   # 데이터 모델 정의
├── api/
│   └── solvedac.go      # solved.ac API 클라이언트
├── scoring/
│   └── calculator.go    # 점수 계산 로직
├── storage/
│   └── storage.go       # 데이터 저장소 관리
├── bot/
│   ├── commands.go      # Discord 명령어 처리
│   ├── competition_handler.go  # 대회 관리 명령어
│   └── scoreboard.go    # 스코어보드 생성
├── errors/
│   └── errors.go        # 중앙화된 오류 관리
├── scheduler/
│   └── scheduler.go     # 자동 스코어보드 스케줄러
├── participants.json    # 참가자 데이터 (실행 시 생성)
└── competition.json     # 대회 데이터 (실행 시 생성)
```

## 라이선스

MIT License