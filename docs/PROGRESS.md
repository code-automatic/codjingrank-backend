# Implementation Progress

Updated: 2026-11-05. This report separates the legacy backend that exists from
the v2 target work; documentation changes do not implement the target items.

## Đã làm được trong backend hiện tại

- Auth JWT/bcrypt, refresh/logout và API Gateway xác thực ở edge.
- Curriculum/problem/admin CRUD theo `Subject → Session → Lesson → Problem`.
- Run đồng bộ qua Piston với sample tests.
- Submit tạo Pending, Redis list queue, worker chạy hidden tests và cập nhật
  verdict; mọi verdict cuối vẫn được lưu theo schema hiện tại.
- `/me/submissions`, `/me/stats`, rate limit/anti-spam và response envelope
  dùng chung.
- PostgreSQL, Redis, Docker Compose, seed/E2E cơ bản và tài liệu OpenAPI hiện tại.

## Chưa làm được / không có trong backend hiện tại

- Chuyển catalog sang `Package/Topic → Level → Problem`, progress/unlock theo
  mô hình mới.
- Leaderboard list theo bài giải; `/me/stats` hiện chỉ là thống kê cá nhân.
- Forum service/API/schema và moderation.
- AI service, `ai_reviews`, rubric/score/findings và gọi tự động sau Accepted.
- Ownership check đầy đủ cho mọi submission và contract test cho API đích.

## Đã loại khỏi flow đích

Repository/GitHub App/OAuth/webhook, repository tables, commit metadata và
post-Accepted commit hook. Code/schema legacy vẫn còn để phục vụ migration;
không gọi chúng là tính năng v2 và không xóa remote repository.

## Lịch tài liệu

06–07/11 schema/catalog/auth; 08–11/11 Run/editor/admin; 12–15/11
Submit/progress/AI schema; 16–19/11 Forum/leaderboard/AI integration; 20–21/11
kiểm thử và sửa lỗi; 22/11 nộp.
