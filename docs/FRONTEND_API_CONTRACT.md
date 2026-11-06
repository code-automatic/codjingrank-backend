# Frontend API Contract (v2.0)

Tài liệu này tách API **đang chạy** khỏi API **đích**. Base URL local: `http://localhost:8080`. Gateway hiện chưa có `/api/v1`; frontend mới nên bọc một client riêng để chuyển prefix sau migration.

Cập nhật 05/11/2026. Mốc triển khai tài liệu: 06/11/2026–22/11/2026. API
Forum, Leaderboard và AI bên dưới chưa implement; không gọi chúng như route live.

## Quy ước chung

Envelope đã chốt ngày 07/11/2026 theo
[DECISIONS.md](DECISIONS.md#api-response-envelope-decision--confirmed-2026-11-07);
xem [response contract](fe/RESPONSE_CONTRACT.md) cho ví dụ đầy đủ.
`code` bằng HTTP status; `error` là chuỗi mã ổn định; `details` luôn có và có
thể `null`. Frontend không rẽ nhánh theo `message`. Verdict và trạng thái AI
nằm trong `data`; `204` không có body. Lần chốt này không đổi pagination,
HTTP mapping hay thêm field envelope.


- JSON dùng `camelCase`; thời gian hiện trả RFC3339.
- Thành công: `{ "status":"success", "code":200, "message":"...", "data":{} }`.
- Lỗi: `{ "status":"error", "code":400, "message":"...", "error":"BAD_REQUEST", "details":null }`.
- API protected dùng `Authorization: Bearer <accessToken>`. Gateway verify HS256 rồi gửi claim nội bộ; frontend không gửi `X-User-ID`.
- Refresh token là cookie HttpOnly `refreshToken`, path `/auth`, TTL mặc định 7 ngày. Access token chỉ giữ trong memory/session layer của client.
- ID hiện tại là số nguyên, trừ `submission.id` là UUID.

## API hiện có (có thể gọi ngay)

### Auth

| Method | Path | Auth | Request / data chính |
|---|---|---|---|
| POST | `/auth/register` | Public | `{email,password,username,fullName}` → access token + user |
| POST | `/auth/login` | Public | `{email,password}` → `{accessToken,user}` |
| POST | `/auth/refresh` | Cookie | rotate cookie → `{accessToken,user}` |
| POST | `/auth/logout` | Public/cookie | revoke cookie |
| GET | `/auth/me` | Bearer | `{user}` |
| GET | `/auth/github/status` | Bearer | GitHub App connection (legacy; do not use in v2) |

Register password tối thiểu 6 ký tự; đây là hành vi code hiện tại, khác SRS yêu cầu 8 ký tự.

### Curriculum/problem

| Method | Path | Auth | Data |
|---|---|---|---|
| GET | `/subjects` | Public | `{subjects:[{id,title,slug,color,languageId}]}` |
| GET | `/subjects/{slug}` | Public | `{subject}` |
| GET | `/subjects/{slug}/sessions` | Public | `{sessions:[...]}` |
| GET | `/sessions/{id}/lessons` | Public | `{lessons:[...]}` |
| GET | `/lessons/{id}` | Public | `{lesson}` |
| GET | `/lessons/{id}/problems` | Public | `{problems:[{id,slug,title,difficulty,orderIndex,acceptanceRate}]}` |
| GET | `/problems/{slug}` | Public | problem detail + `sampleTestCases`; không trả `driverCode` |
| GET | `/problems/{id}/test-cases` | Public | samples; gateway ép `visibility=sample` |
| GET | `/problems/daily-challenge` | Public | daily challenge; endpoint có thể tạo row cho ngày hiện tại |

### Run và Submit

```json
// POST /run (Bearer)
{ "problemId": 12, "languageId": 2, "code": "print(1)" }
// 200
{ "status":"success", "code":200, "message":"Code executed successfully", "data":
  { "status":"Accepted", "passed":2, "total":2,
    "results":[{"testCaseId":1,"status":"Accepted","inputData":"...",
      "expectedOutput":"1","actualOutput":"1","runtimeMs":12}] } }
```

`/run` là synchronous, chạy sample và không có run ID. Request code tối đa 100 KB ở DTO; SRS 64 KB chưa đúng với code.

```json
// POST /submit (Bearer)
{ "problemId": 12, "languageId": 2, "code": "print(1)" }
// 201 -> data.submission: { id, userId, problemId, languageId, code, status,
// runtimeMs, memoryKb, errorMessage, passCount, totalTestcases, createdAt, judgedAt }
```

Submit tạo Pending ngay lập tức, queue Redis xử lý sau. Frontend hiện có thể poll `GET /submissions/{id}`, nhưng endpoint này cần backend bổ sung ownership check trước khi coi là an toàn. Các endpoint lịch sử là `GET /me/submissions`, `GET /me/submissions/{problemId}/latest`; thống kê là `GET /me/stats`.

### GitHub/repository (legacy, sẽ loại bỏ)

Các route `/repositories*`, `/github/*` và webhook/callback đang tồn tại trong
backend cũ nhưng **không thuộc sản phẩm đích**. Frontend mới không tích hợp
chúng. Migration sẽ xóa repo/GitHub tables, DTOs, routes và commit metadata.

### Admin

JWT role admin mới được qua `/admin/*`. Các nhóm hiện có: `/admin/languages`, `/admin/subjects`, `/admin/subjects/{id}/sessions`, `/admin/sessions/{id}`, `/admin/sessions/{id}/lessons`, `/admin/lessons/{id}`, `/admin/problems`, `/admin/problems/{id}`, `/admin/problems/{problemId}/test-cases`, `/admin/tags`, và liên kết problem-tag/lesson-problem. Đây là CRUD theo schema cũ, chưa phải CRUD Topic/Level của SRS.

## API đích cần thêm

| Boundary | Sở hữu | Trạng thái |
|---|---|---|
| `/api/v1/forum/*` | `forum_service` | Chưa implement |
| `/api/v1/leaderboard*` | Core/Leaderboard module | Chưa implement |
| `/api/v1/submissions/{id}/ai-review` | `ai_service` đọc qua Gateway | Chưa implement |

Mọi response đích vẫn bọc trong `data` theo envelope chung; các ví dụ dưới
chỉ mô tả payload bên trong `data`.

### Forum (service mới, chưa implement)

Forum là service độc lập sau API Gateway. Nó không đọc GitHub và không chạy
trong judge worker. Các route learner đề xuất:

```text
GET    /api/v1/forum/posts?problemId=&postType=&sort=&page=1&size=20
POST   /api/v1/forum/posts
GET    /api/v1/forum/posts/{id}
PATCH  /api/v1/forum/posts/{id}
DELETE /api/v1/forum/posts/{id}
GET    /api/v1/forum/posts/{id}/comments
POST   /api/v1/forum/posts/{id}/comments
PATCH  /api/v1/forum/comments/{id}
DELETE /api/v1/forum/comments/{id}
PUT    /api/v1/forum/posts/{id}/like
DELETE /api/v1/forum/posts/{id}/like
```

Post tối thiểu gồm `postType` (`discussion|question|solution`), `title`,
`contentMd`, `codeSnippet` tùy chọn, `problemId`/`languageId` tùy chọn. Comment
chỉ lồng một cấp. Server phải sanitize Markdown, kiểm tra ownership và chặn
`solution` nếu user chưa giải problem theo progress mới.

### Catalog sau migration

```text
GET /api/v1/packages
GET /api/v1/packages/{slug}
GET /api/v1/packages/{slug}/levels
GET /api/v1/problems?package={slug}&level=easy&solved=true&page=1&size=20
GET /api/v1/problems/{slug}
GET /api/v1/problems/{slug}/template?languageId=2
```

Đề xuất response level: `{package, levels:[{level,difficulty,total,solved,unlocked,completed}]}`. Server phải tính unlock, không tin cờ từ client. Migration cần mapping dữ liệu cũ: `subjects` → package, các lesson/session → metadata hoặc bỏ khỏi navigation, `problems.difficulty` → level.

### Leaderboard (không dùng AI score)

```text
GET /api/v1/leaderboard?scope=all|week|month&package={slug}&page=1&size=50
GET /api/v1/leaderboard/me
```

Response đề xuất:

```json
{ "scope":"all", "items":[
  { "rank":1,"userId":7,"username":"an","avatarUrl":null,
    "solvedProblemCount":34,"acceptedSolutionCount":41 }
], "page":1,"size":50,"total":1280 }
```

`solvedProblemCount` là số problem Accepted distinct; `acceptedSolutionCount` là số Accepted. AI review không làm thay đổi hai trường này.

### AI code review (đánh giá riêng, chưa implement)

AI thay cho post-judge repository commit hook. Judge vẫn chạy và quyết định
verdict như hiện tại; AI chỉ đánh giá chất lượng source, không thay verdict và
không tham gia leaderboard. Đã chốt: worker tự động gọi AI sau khi lưu `Accepted`.

Run và mọi verdict khác Accepted không kích hoạt AI. Frontend đọc kết quả AI
riêng; không chờ AI để hiển thị Accepted. Lỗi AI không thay đổi verdict hoặc rank.

```text
GET  /api/v1/submissions/{submissionId}/ai-review
```

Frontend không cần POST để kích hoạt AI. Đề xuất POST thủ công trước đây được
rút khỏi phạm vi; cơ chế retry/review lại chưa chốt. Backend lấy code, language
và problem từ submission đã Accepted, không nhận bản thay thế do client gửi.

Payload trong `data` đề xuất: `{ "status":"queued|running|completed|failed", "score":0..100,
"summary":"...", "dimensions":{"correctness":0,"complexity":0,"readability":0,
"maintainability":0}, "findings":[{"severity":"info|warning|error","line":3,
"title":"...","explanation":"...","suggestion":"..."}], "model":"...",
"createdAt":"...","completedAt":"..." }`.

Backend phải kiểm tra owner, giới hạn quota, không gửi hidden test/secret/provider key cho client, lưu model/prompt version để tái lập kết quả, và coi AI failure là trạng thái riêng không làm bài submit thành Internal Error. AI service nên nhận submission ID, source, language và problem context tối thiểu qua service-to-service auth; không nhận GitHub/repository data.

## Frontend migration rule

Frontend hiện tại dùng các route gốc và DTO số nguyên. Frontend mới dùng một API client có `BASE_PATH` cấu hình được, model `Package/Level`, và adapter tạm thời; không tự suy ra level từ `difficulty` hoặc coi `rating` trong `/me/stats` là leaderboard score. Chỉ bật màn leaderboard/AI khi endpoint backend tương ứng đã được triển khai và có contract test.
