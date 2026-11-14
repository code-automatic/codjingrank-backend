# Backend Reality Audit

> **Current baseline 14/11/2026:** Repository/GitHub, the old hierarchy and old
> migrations are removed. UUID catalog/admin/Run/Submit are integrated. Sprint 3
> adds root/v1 submission detail/history/latest/stats, active owner/admin reads,
> bounded worker recovery with attempt fencing, and transactional Accepted
> solutions/progress. Trusted real-Piston lost-enqueue smoke passed. Full v1
> Auth/profile, Forum, AI and leaderboard APIs remain pending. See
> [Sprint 3 acceptance](SPRINT3_ACCEPTANCE.md) and [project state](../PROJECT_STATE.md).

## Historical SRS audit — 05/11/2026

The inventory below is the original audit snapshot. Its references to Repo,
integer IDs, legacy stats/rating and missing target-schema tables are historical,
not current runtime claims. Current evidence above supersedes those statements.


**Đối chiếu:** backend trong repository này với SRS mới tại `code-automatic/srs` (commit `a46397729ca56e74bf6dabf057c6ce82f102b66b`, 2026-11-05). Đây là snapshot phân tích; chưa phải cam kết rằng mọi chức năng trong SRS đã được triển khai.

## Kết luận nhanh

Backend hiện là MVP RinnoGen chạy được: Go monorepo, API Gateway `:8080`, Auth `:8081`, Run `:8082`, Submission `:8083`, Repo `:8084`; PostgreSQL, Redis và Piston trong Docker Compose. Luồng chạy code đã có; luồng submit bất đồng bộ đã có. **Leaderboard dạng danh sách chưa có. AI code review/chấm chất lượng chưa có.** `GET /me/stats` chỉ trả hạng cá nhân và rating công thức đơn giản. `repo_service` hiện chỉ phục vụ GitHub/repository/commit; theo quyết định mới, service này sẽ bị loại khỏi sản phẩm.

SRS mới mô tả một hệ thống khác ở vài ranh giới quan trọng: gói chủ đề/cấp độ, UUID và bảng dữ liệu mới, lưu chỉ Accepted, API `/api/v1`, Core/Judge tách khác, Asynq/sqlc/Goose, diễn đàn và leaderboard. Không được dùng SRS làm mô tả hiện trạng.

## Conflict với code thực tế

| Chủ đề | SRS/tài liệu nói | Code hiện tại | Mức độ |
|---|---|---|---|
| Cấu trúc học tập | `problem_categories` → 3 level | `subjects` → `subject_sessions` → `lessons` → `lesson_problems`; `problems.difficulty` chỉ là Easy/Medium/Hard | **Breaking** |
| ID và schema | UUID, `problem_categories`, `user_problem_progress` | ID người dùng/bài/ngôn ngữ là `int`; `schema.sql` không có category/progress/leaderboard view | **Breaking** |
| Prefix API | `/api/v1/...` | Gateway dùng đường dẫn gốc: `/subjects`, `/problems/{slug}`, `/run`, `/submit` | **Breaking** |
| Run | 202 + job/polling, Redis TTL | `POST /run` chạy đồng bộ, trả 200 và kết quả sample trong response; không có `GET /runs/{id}` | **Breaking** |
| Submit persistence | chỉ lưu Accepted, lỗi không vào DB | tạo `Pending` ngay khi nhận, worker cập nhật mọi verdict (`Accepted`, `Wrong Answer`, `Compilation Error`, ...) vào `submissions` | **Breaking** |
| Submit response | `job_id`, polling job | 201 trả toàn bộ submission với UUID; worker không có API trạng thái job riêng | **Breaking** |
| Test ẩn | không trả cho learner | API gateway chỉ expose sample, nhưng submission worker đọc **mọi** test case; cần giữ boundary này | Rủi ro |
| Publish/level lock | kiểm tra publish và mở khóa | Run/Submit kiểm tra tồn tại bài/ngôn ngữ; chưa có level unlock và Run không lọc `is_published` | **Missing** |
| Verdict | snake_case (`accepted`, `wrong_answer`...) | enum/title case (`Accepted`, `Wrong Answer`...) | Contract |
| Rate limit | riêng Run 10/phút, Submit 5 giây | gateway rate limit chung theo IP; DB anti-spam Submit 10 giây/user+problem | Khác |
| Auth | RS256/JWKS, `/api/v1/users/me` | HS256 với `JWT_SECRET`, `/auth/me`, refresh cookie HttpOnly | Security/Contract |
| Curriculum API | categories/problems filter/pagination | subjects/sessions/lessons; problem list chỉ theo lesson | **Breaking** |
| Leaderboard | `GET /leaderboard`, scope/category/cache | Không có route; `GET /me/stats` tính `currentRank`, `totalUsers`, `rating = 1200 + solved*10` | **Missing** |
| AI scoring | SRS nguồn v1 ghi ngoài phạm vi; v2 đã yêu cầu AI sau Accepted | Không có package, bảng, queue hay provider AI | **Missing** |
| Forum | 12 chức năng forum | Không có bảng/route forum trong `schema.sql` hoặc gateway | **Missing** |
| Tooling | chi/sqlc/Goose/Asynq | `net/http`, pgx trực tiếp, SQL thủ công, Redis list `LPush/BRPop`; không có các dependency SRS nêu | Docs stale |
| Sandbox | Docker hardening chi tiết | Piston container `privileged: true`, image `latest`; cần security review trước production | **High risk** |

## Ranh giới triển khai mới đã chốt

| Phần | Quyết định |
|---|---|
| Judge hiện tại | Giữ nguyên cơ chế Run sample, Submit, Redis queue, worker, Piston, verdict và rate-limit/anti-spam trong giai đoạn migration. Không viết lại thành Asynq hay service judge mới chỉ vì SRS mô tả như vậy. |
| Schema | Thay schema để biểu diễn Package/Level, progress, leaderboard, forum và AI review; cập nhật query/DTO theo schema mới. Schema migration không được tự động kéo theo thay đổi verdict hoặc cách worker chạy code. |
| Repo/GitHub | Loại bỏ `repo_service`, GitHub App/OAuth/webhook, bảng repository/GitHub, commit metadata và best-effort commit sau Accepted. Đây là phần bị bỏ, không phải phần tạm hoãn. |
| AI thay repo hook | Tại điểm worker hiện gọi `POST /internal/commits/accepted-submission`, gọi AI service nội bộ để review source. Judge vẫn quyết định `Accepted/WA/CE/...`; AI chỉ tạo score, summary và findings, không sửa verdict và không tham gia leaderboard. |
| Forum | Bổ sung `forum_service` và schema/handler độc lập: post, comment, like, moderation và chống XSS theo SRS. Không trộn forum vào worker chấm code. |
| AI service | Bổ sung `ai_service` nội bộ, chỉ nhận request từ Submission Worker hoặc API backend có service auth; không expose provider key hay endpoint provider trực tiếp cho frontend. |
| Leaderboard | Tính từ Accepted problem distinct và số Accepted; không đọc AI score. |

### Luồng mục tiêu sau khi bỏ repository

```text
Frontend -> API Gateway -> Submission API
  -> PostgreSQL: tạo Pending
  -> Redis: enqueue
  -> Judge Worker -> Piston -> cập nhật verdict
  -> chỉ khi verdict = Accepted: gọi AI Service nội bộ -> lưu ai_reviews
  -> Frontend đọc submission + ai-review
```

Đã chốt: AI tự động chạy chỉ sau khi lưu verdict `Accepted`, thay trực tiếp post-judge commit hook hiện tại. Run và các verdict khác không kích hoạt AI. Lỗi AI không đổi verdict Accepted hoặc điểm xếp hạng. Repo flow sẽ bị loại bỏ; hiện chỉ cập nhật tài liệu, chưa xóa code/schema hay repository trên GitHub.

## Những gì đã xác minh

- `api_gateway/internal/route/route.go` là hợp đồng route thật; Gateway proxy inject `X-User-ID`, `X-User-Email`, `X-User-Role` sau khi verify JWT.
- `run_service/internal/service/service.go` đọc sample từ `test_cases`, ghép `driver_code` cho `function`, gọi Piston và trả từng test result đồng bộ.
- `submission_service/internal/service/service.go` insert Pending và enqueue Redis; `submission_service/internal/worker/worker.go` chạy hidden/sample theo query hiện tại, cập nhật verdict và best-effort commit GitHub.
- `submission_service/internal/repository/stats.go` đếm distinct Accepted problem, streak, rank cá nhân; chưa có bảng xếp hạng nhiều người.
- `shared/response` chuẩn hóa envelope `{status, code, message, data}` và lỗi `{status, code, message, error, details}`.

## Hướng đích đã chốt

1. Chuyển catalog sang **Topic/Package → Level (Easy/Medium/Hard) → Problem**. Đây là migration/adapter lớn; không đổi tên `subjects` thành category bằng find-and-replace.
2. Giữ nguyên logic chạy/chấm code hiện tại; chỉ đổi schema, repository/query và DTO cần thiết để chứa mô hình mới.
3. Xóa repo/GitHub flow; thay post-judge commit hook bằng AI service call nội bộ. AI là **đánh giá riêng**, không thay verdict và không tham gia leaderboard.
4. Bổ sung forum và leaderboard bằng service/API mới, không đưa chúng vào judge worker.
5. Giữ API hiện tại cho client cũ trong giai đoạn chuyển tiếp; xây `/api/v1` façade sau khi chốt migration schema. Không để frontend mới gọi các endpoint SRS chưa tồn tại.

## Ưu tiên triển khai

**P0:** ownership khi đọc `/submissions/{id}`, chặn bài chưa publish, sandbox hardening, job recovery/idempotency, và contract test. Policy AI đã chốt: chỉ Accepted.
**P1:** migration Topic/Level, Accepted/progress model, loại repo/GitHub schema và flow, OpenAPI v2.
**P2:** leaderboard read API, forum schema/API, AI review service/queue và quota/cost control.
**P3:** daily challenge mở rộng, analytics.

## Phạm vi tài liệu

SRS mới, backlog và kế hoạch tích hợp là tài liệu **đề xuất**; `schema.sql`, route Go và DTO Go là nguồn sự thật hiện tại. Khi triển khai migration, cập nhật OpenAPI, seed, frontend client và acceptance tests cùng một thay đổi.
