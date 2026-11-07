# Quy trình phối hợp và tích hợp — v2.0

Ngày cập nhật: 05/11/2026. Thực hiện: 06/11/2026–22/11/2026. Đây là quyết định thiết kế và kế hoạch, chưa phải triển khai code.

Flow sản phẩm đích giữ Run/Submit/Judge/Piston/Redis và loại bỏ toàn bộ
Repository/GitHub/commit flow. Sau khi worker ghi `Accepted`, worker gọi
`ai_service` nội bộ; AI review không ảnh hưởng verdict hoặc leaderboard. Forum
là service riêng, không chạy trong judge worker.


## Contract mock frontend — cập nhật 07/11/2026

Nguồn API target duy nhất: [OpenAPI đầy đủ](../openapi.yaml) (94 operation, có
admin) và [learner projection](../openapi_fe.yaml) (42 operation). Hướng dẫn
mock/UUID/schema gaps tại [frontend contract](../FRONTEND_API_CONTRACT.md).
Mọi operation target-not-implemented; API đã có route root vẫn cần tích hợp
UUID/schema và prefix v1. Không dùng route Subject/Session/Lesson/GitHub đã xóa.
Database mới dùng chung, DATABASE_URL; không tách Auth/Judge database. Các
đoạn kế hoạch gốc dưới đây chỉ tham khảo khi mâu thuẫn baseline mới. Optional
password-reset vẫn cần token store/email; avatar/bio đã có cột schema nhưng Go
runtime chưa tích hợp, không coi mock là đã làm BE.

## 1. Owner và hỗ trợ

| Người | Owner chính | Reviewer / kiểm tra tích hợp |
|---|---|---|
| TV1 — Quốc Thịnh | Auth, common-auth/errors, Forum API/UI, Profile, test plan, PO | Review Core của TV3; xác nhận tiêu chí nghiệm thu |
| TV2 — Tùng | Run/Judge/worker/sandbox, Leaderboard, AI trigger, SM | Review Auth TV1; FE gọi Run/Submit/AI TV4; bảo mật execution |
| TV3 — Hiếu | Core/catalog/levels/history, common-db/levels, migration/seed; AI review schema | Review TV2 về DB/transaction; Compose/CI và cấu hình chung |
| TV4 — Huy | Next.js, shared UI, editor, Admin, các trang trừ Forum/Profile | Review UI của TV1; contract frontend/backend |

**Hỗ trợ được phân công:** Hiếu (TV3) hỗ trợ Tùng (TV2) về DB/transaction, shared package và tích hợp Core–Judge; Quốc Thịnh (TV1) hỗ trợ Huy (TV4) về Auth, phân quyền, API user và tích hợp UI. Owner chính giữ trách nhiệm đầu ra và cập nhật trạng thái task. Người tham gia viết phần thay đổi cần một thành viên khác review trước khi hoàn thành.

AI hỗ trợ toàn dự án nhưng mỗi lượt làm bám một task/owner cụ thể: đọc yêu cầu, kiểm tra phụ thuộc, triển khai trong phạm vi, kiểm thử, đưa kết quả review. Không tự nhận task của thành viên khác là đã hoàn tất. Trước sửa shared package hoặc config phải kiểm tra các bên dùng nó. Nếu phát hiện xung đột, báo mã task, tệp/hợp đồng bị ảnh hưởng, owner và cách xử lý. Không ghi đè thay đổi chưa commit của người khác. Không đóng vai reviewer con người hoặc PO nghiệm thu thay team.

## 2. Hợp đồng phải thống nhất

- OpenAPI theo service, request/response/error/pagination được version cùng PR. Mock FE dựa đúng hợp đồng; client type sinh từ OpenAPI. API thay đổi phải cập nhật cả mock, client và acceptance tests liên quan.
- Một Go module, một bộ phiên bản backend; frontend một package.json và lockfile. CI kiểm tra gofmt, go vet, golangci-lint, go test, TypeScript, lint và next build. Go test -race cho logic concurrency trên runner hỗ trợ; PostgreSQL/Redis thật cho integration tests.
- Biến môi trường thống nhất trong .env.example không có secret: AUTH_DATABASE_URL, JUDGE_DATABASE_URL, REDIS_URL, JWT_ISSUER, JWT_AUDIENCE, JWT_PRIVATE_KEY_PATH, JWT_PUBLIC_KEY_PATH, INTERNAL_SERVICE_TOKEN, AUTH_INTERNAL_URL, CORE_INTERNAL_URL, RUN_INTERNAL_URL, JUDGE_INTERNAL_URL, WEB_ORIGIN. NEXT_PUBLIC_* chỉ chứa dữ liệu công khai; tuyệt đối không DB URL/secret.
- Port dev nội bộ đề xuất: Next 3000; Auth 8081; Core 8082; Run 8083; Judge API 8084; PostgreSQL 5432; Redis 6379. Docker dùng service DNS, không localhost để liên lạc giữa container. Chỉ Next public ở môi trường demo, backend giữ mạng nội bộ.
- Gateway v1 dùng prefix `/api/v1` làm canonical public contract và strip prefix khi forward; `GET /api/v1/users/me` rewrite riêng tới Auth `/auth/me`. Chỉ map các auth route hiện có, không map GitHub auth. Run: `POST /api/v1/run` → Run; Submit/history/stats → Submission; catalog/admin legacy → Gateway; Forum/Leaderboard/AI là target sau migration. Giữ root aliases trong migration, không map `/internal/*` ra public. Bảo toàn `Set-Cookie`, `Origin`, request ID; Bearer chỉ dùng Gateway xác thực và không forward sang upstream protected. Test cookie path khi bật v1. Run vẫn synchronous, Submit vẫn Pending → Redis → worker → verdict.
- TV3 quản lý migration Judge, TV1 Auth. Migration đã merge không sửa lại; thêm migration mới. Unique/check/FK baseline giữ nguyên. Test dựng DB sạch và nâng cấp DB có dữ liệu trước merge.
- Sandbox nhận input qua stdin, không mount DB/secret/output mong đợi. Giới hạn CPU/memory/pids/file/output, network none, non-root, read-only, seccomp, cleanup. Java có cấu hình JVM phù hợp memory limit.
- Job và run có user_id, chỉ chủ sở hữu xem; queue giao lại không tạo AC trùng. Giới hạn timeout, retry, TTL, concurrency phải dùng config chung và kiểm thử khi Redis/worker lỗi.

## 3. Lịch và cổng nghiệm thu

| Ngày | TV1 — Quốc Thịnh | TV2 — Tùng | TV3 — Hiếu | TV4 — Huy | Điều kiện tích hợp |
|---|---|---|---|---|---|
| 06–07/11 | Auth contract, middleware, Forum/AI policy | Sandbox/queue contract, Python runner | Package/Level schema, migration baseline | Next.js/TS, routes/layout, mock contract | Khung chạy Compose; API live/đích được đánh dấu |
| 08–11/11 | Admin user/test plan; soạn bài | C++/wrapper, Run, Judge API/Redis | Catalog/template/publish guard, seed | Monaco/gói/bài/Run; Admin nội dung | Run Python/C++; không lộ hidden tests |
| 12–15/11 | Forum CRUD, hồ sơ | Java, verdict, progress/level, AI trigger contract | History/ownership, `ai_reviews` schema, seed | Submit/polling/history, AI status placeholder | Submit giữ cơ chế cũ; Accepted mới tạo AI job |
| 16–19/11 | Forum comments/likes/moderation/XSS và UI | Leaderboard/cache, AI service integration, E2E | Leaderboard query, AI repository, audit Core | Leaderboard, Forum/AI UI, responsive | Must đạt; đóng tính năng 19/11 lúc 21:00 |
| 20–21/11 | PO nghiệm thu, tài liệu/demo | Sandbox attack/load/regression, AI failure path | Security/migration sạch/seed/backup | Cross-browser/build/regression, API contract | CI xanh, không lỗi nghiêm trọng; chạy demo từ DB sạch |
| 22/11 | Nộp dự án trước giờ giảng viên quy định | Trực hỗ trợ Judge | Trực DB/deploy | Trực frontend | Không thêm tính năng mới |

Giữ 67 task và 72 story; phân bổ chi tiết ngày trước đó vẫn dùng, thay các đầu ra Python backend bằng Go tương ứng. Không thêm runtime Go cho bài nộp. Nếu 11/11 Run chưa đạt hoặc 15/11 AC/levels chưa đạt: dừng Should, dồn hỗ trợ vào Must và báo PO; không hạ chuẩn sandbox để chạy kịp. Chưa biết số giờ/ngày nên lịch là mục tiêu, không phải bảo đảm khả thi.

## 4. Trước và sau mỗi task

1. Đọc SRS v2.0 và acceptance story; kiểm tra git status, nhánh, owner; không overwrite việc dở của người khác.
2. Kiểm tra task chặn: API đã có hay mock, package/config/migration đã merge chưa; ghi blocker cụ thể.
3. Thực hiện task trên feature/<task-id>-<slug>; tránh đổi dependency/contract ngoài phạm vi không có lý do.
4. Kiểm thử logic task và integration bị ảnh hưởng. FE đổi phải build; DB đổi phải migrate trên DB sạch; shared package đổi phải test các consumer.
5. PR ghi task, thay đổi, validation, breaking changes và owner bị ảnh hưởng; một người khác review. Chỉ tick khi merge/CI/PO đạt DoD.

Các workbook/SRS nguồn là tài liệu tham chiếu, không thay thế code hiện tại. Khi giao task phải bám Go `net/http`, pgx, Redis list queue, Piston và Next.js/TypeScript; không thêm Asynq, chi, Goose hoặc repo/GitHub flow chỉ vì tài liệu nguồn cũ ghi như vậy.
