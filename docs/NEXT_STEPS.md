# Next Steps — v2.0

> **Cleanup 06/11/2026:** Repository/GitHub and Subject/Session/Lesson code,
> old migrations/seeds and the Vite test UI have been removed on
> `chore/remove-legacy-code`. Any inventory of those implementations below is
> historical. Retained Go persistence/DTOs still require UUID/schema integration;
> `/api/v1`, Forum and AI remain target-only. See [project state](../PROJECT_STATE.md).


Updated: 2026-11-05. Window: 06/11/2026–22/11/2026.

## P0 — giữ luồng judge

- Giữ Run sample, Submit Pending → Redis → worker → Piston → verdict.
- Bổ sung ownership check cho `GET /submissions/{id}`, job recovery/idempotency
  và contract tests.
- Không để AI lỗi biến Accepted thành Internal Error.

## P1 — tích hợp schema và catalog

- Schema mới và legacy cleanup đã có; không migrate dữ liệu hierarchy cũ.
- Tích hợp Go DTO/JWT/query sang UUID và packages/levels trong DB dùng chung.
- Hoàn thiện API theo [OpenAPI target](openapi.yaml); frontend mock theo
  [learner spec](openapi_fe.yaml), không coi đây là route live.
- Triển khai ownership, progress/unlock và leaderboard Accepted distinct;
  kiểm thử transaction/recovery cùng DB/worker.

## P2 — tính năng mới

- Forum service: posts, comments, likes, moderation, chống XSS.
- Leaderboard read API theo bài giải; không đọc AI score.
- AI service: chỉ nhận job nội bộ sau Accepted, lưu `ai_reviews`, trả trạng thái
  queued/running/completed/failed và score/findings riêng.
- Cập nhật Gateway, frontend client và OpenAPI target cùng contract tests.

## Lịch

| Ngày | Cổng nghiệm thu |
|---|---|
| 06–07/11 | schema/catalog/auth và contract |
| 08–11/11 | Run/editor/admin |
| 12–15/11 | Submit/progress/AI schema |
| 16–19/11 | Forum/leaderboard/AI integration; đóng tính năng 19/11 21:00 |
| 20–21/11 | E2E, security, regression, tài liệu |
| 22/11 | Nộp/demo; không thêm tính năng |

Các lệnh E2E/GitHub trong tài liệu cũ là legacy và không phải tiêu chí v2.
