# BẢNG PHÂN CÔNG NHIỆM VỤ

## Dự án: Coding Judge Platform (team 4 thành viên)

| | |
|---|---|
| **Phiên bản** | 2.0 (khớp SRS 2.0 và backlog 2.0) |
| **Ngày cập nhật** | 05/11/2026 |
| **Trạng thái** | Kế hoạch đích; chưa phải trạng thái triển khai |
| **Tài liệu gốc** | `01_SRS_Coding_Judge_Platform.md` |
| **Tài liệu đi kèm** | `03_BACKLOG_THEO_SPRINT.md` |

**Cách đọc tài liệu:** mã `FR-xx.y` trỏ về yêu cầu trong SRS; mã `Tn-xx` là nhiệm vụ của thành viên TVn (ví dụ `T2-05` là nhiệm vụ số 5 của TV2). Nhiệm vụ nào cần thành viên khác cung cấp trước được ghi ở cột **Phụ thuộc**.

**Thay đổi so với bản 5 người (1.3):** vai trò TV5 (Social + QA) được chia lại cho TV1, TV2, TV3: Diễn đàn và kiểm thử do TV1 đảm nhận, Bảng xếp hạng do TV2 (sau khi xong lõi Judge), kiểm thử end-to-end và bảo mật chia cho TV2 và TV3. Một số giao diện chuyển từ TV4 sang TV1 để cân tải.

*Giả định: team 4 người làm theo Scrum trong cửa sổ 06/11/2026–22/11/2026. Các nhiệm vụ mức **Could** sẽ cắt trước khi trễ. Run/Submit/Judge giữ cơ chế cũ; repo/GitHub không còn là nhiệm vụ sản phẩm.*

---

## 1. THÀNH VIÊN VÀ VAI TRÒ

| TV | Họ tên | Vai trò Scrum | Vai trò kỹ thuật | Sở hữu (owner) |
|---|---|---|---|---|
| TV1 | Quốc Thịnh | Product Owner | Backend (Auth, Forum) + Frontend (Forum, Hồ sơ) | Auth Service, Forum API, `common-auth`, `common-errors`, test plan |
| TV2 | Tùng | Scrum Master | Judge Engine + Leaderboard + AI integration | Run Service, Submission Worker, sandbox, Leaderboard API, gọi AI sau Accepted |
| TV3 | Hiếu | Developer | Backend (Core Problem) | Core API (nội dung bài), `common-db`, `common-levels`, migration, seed |
| TV4 | Huy | Developer | Frontend | Ứng dụng Next.js + TypeScript (trừ trang Diễn đàn và Hồ sơ) |

**Phân công hỗ trợ:** Hiếu (TV3) hỗ trợ Tùng (TV2); Quốc Thịnh (TV1) hỗ trợ Huy (TV4). Tùng và Huy vẫn là owner chính các task được giao; giữ nguyên mã TV và mã nhiệm vụ để truy vết.

- **Hiếu → Tùng:** hỗ trợ DB/query/transaction, tích hợp `common-db`, `common-levels` và xử lý xung đột cấu hình giữa Core và Run/Judge; ưu tiên bảo đảm tiến độ Core.
- **Quốc Thịnh → Huy:** hỗ trợ tích hợp đăng nhập/refresh/phân quyền, API quản trị user, component dùng chung với Forum/Profile và kiểm thử luồng frontend/backend.
- Người hỗ trợ ghi rõ phần đã làm trong task/PR; chỉ tick hoàn thành khi đủ đầu ra, kiểm thử và review theo DoD.

**Git Lead: Hiếu (TV3)**, Tùng (TV2) hỗ trợ cấu hình execution worker. Phụ trách quy trình nhánh, quy ước commit, bảo vệ nhánh `main` và `develop`.

**Về vai trò Scrum Master:** TV2 làm SM nhưng cũng là người có khối lượng kỹ thuật lớn nhất ở Sprint 1–3. Nếu team thấy quá tải, có thể giao SM cho TV1 hoặc TV3 (nghi thức Scrum chỉ khoảng 3–4 giờ mỗi sprint). Đổi vai trò thì cập nhật cột "Vai trò Scrum" và nhiệm vụ T2-15.

---

## 2. MA TRẬN CHỨC NĂNG VÀ NGƯỜI PHỤ TRÁCH

R = chịu trách nhiệm chính (backend/logic). S = hỗ trợ. Cột TV4 là phần giao diện, trừ các ô ghi rõ.

| Chức năng | TV1 | TV2 | TV3 | TV4 |
|---|---|---|---|---|
| F01 Đăng ký / Đăng nhập | R (API) | | | S (UI) |
| F02 Landing page | S (user count) | | S (API landing) | R |
| F03 Chọn ngôn ngữ | | S | R | S (UI) |
| F04 Chọn gói bài | | | R | S (UI) |
| F05 Code mẫu | | S (wrapper) | R | S (UI) |
| F06 Run code | | R | S | S (UI) |
| F07 Submit | | R | S | S (UI) |
| F08 Level up | | S | R | S (UI) |
| F09 Lịch sử nộp bài | | | R | S (UI) |
| F10 Bảng xếp hạng | S (user lookup) | R (API) | | S (UI) |
| F11 Diễn đàn | R (API + UI) | | | S (component dùng chung) |
| F13 AI review | S (policy) | R (trigger/integration) | S (schema/repository) | S (UI đọc kết quả) |
| F12 Quản trị | R (user, kiểm duyệt diễn đàn) | | R (nội dung) | R (UI) |

---

## 3. NHIỆM VỤ CHI TIẾT THEO THÀNH VIÊN

### TV1 — Quốc Thịnh: Product Owner, Auth Service, Diễn đàn

| Mã | Nhiệm vụ | FR | Sprint | Đầu ra | Phụ thuộc |
|---|---|---|---|---|---|
| T1-01 | Thiết kế Auth DB (users, refresh_tokens, password_resets) và migration SQL | FR-01.1 | 0 | Migration Auth DB theo tool của repository | |
| T1-02 | API đăng ký, đăng nhập, hash mật khẩu, cấp JWT | FR-01.1–01.3 | 1 | `/auth/register`, `/auth/login` chạy được | T1-01 |
| T1-03 | Refresh token, đăng xuất, thu hồi token | FR-01.4 | 1 | `/auth/refresh`, `/auth/logout` | T1-02 |
| T1-04 | Package `common-auth`: middleware xác thực JWT, `RequireAdmin` | FR-01.5 | 1 | Go package import được trong cùng module; `go test ./...` pass | T1-02 |
| T1-05 | Package `common-errors`: lỗi và handler theo SRS 2.3 | SRS 2.3 | 0 | Package dùng chung | |
| T1-06 | API nội bộ tra cứu hàng loạt user (có service token) | FR-01.9 | 2 | `/internal/users?ids=` | T1-02 |
| T1-07 | API admin: danh sách, tìm kiếm, khóa/mở khóa, đổi vai trò | FR-01.6, FR-12.1 | 2 | `/admin/users` | T1-04 |
| T1-08 | Hồ sơ cá nhân, đổi mật khẩu | FR-01.8 | 3 | `/users/me` | T1-02 |
| T1-09 | Quên mật khẩu qua email | FR-01.7 | 3–4 | `/auth/forgot-password`, `/reset-password` | T1-02 |
| T1-10 | Giới hạn đăng nhập sai | FR-01.10 | 3–4 | Rate limit Redis cho login | T1-02 |
| T1-11 | Quản lý backlog, viết tiêu chí chấp nhận, chủ trì Sprint Review, nghiệm thu | | 0–5 | Backlog cập nhật, biên bản nghiệm thu | |
| T1-12 | Duy trì SRS (version, changelog) | | 0–5 | SRS cập nhật sau mỗi sprint | |
| T1-13 | Forum API: bài viết (tạo, sửa, xóa mềm, danh sách, lọc, sắp xếp, tìm theo tiêu đề) | FR-11.1–11.3, FR-11.6, FR-11.9 | 3–4 | `/forum/posts` | T3-02, T1-04 |
| T1-14 | Forum API: bình luận lồng một cấp, like, `view_count` | FR-11.4–11.5, FR-11.7 | 4 | `/forum/.../comments`, `/like` | T1-13 |
| T1-15 | Chống spoiler lời giải; render Markdown an toàn (chống XSS) | FR-11.8, FR-11.10 | 4 | 403 `SOLUTION_NOT_ALLOWED`, test XSS | T1-13 |
| T1-16 | Admin diễn đàn: ẩn/xóa bài viết và bình luận | FR-12.9 | 4 | `/admin/forum/...` | T1-13 |
| T1-17 | Test plan, bộ test case theo từng chức năng, mẫu báo lỗi | | 1–2 | Tài liệu test plan | |
| T1-18 | Tài liệu hướng dẫn sử dụng và báo cáo kiểm thử cuối kỳ | | 5 | Hai tài liệu cuối kỳ | T2-19, T3-16 |
| T1-19 | Giao diện Diễn đàn: danh sách, chi tiết, bình luận lồng, like, Markdown an toàn | FR-11 | 4 | Trang diễn đàn | T1-14 |
| T1-20 | Giao diện Hồ sơ cá nhân | FR-01.8 | 4 | Trang hồ sơ | T1-08 |

### TV2 — Tùng: Scrum Master, Judge Engine, Bảng xếp hạng

| Mã | Nhiệm vụ | FR | Sprint | Đầu ra | Phụ thuộc |
|---|---|---|---|---|---|
| T2-01 | Image Docker Python 3; script chạy giới hạn CPU/RAM/time/process, không mạng | NFR-03 | 1 | Image + script chạy một file với input | |
| T2-02 | Image C++17 | NFR-03 | 2 | Image + biên dịch và chạy | T2-01 |
| T2-03 | Image Java 17 (dời sang Sprint 4 nếu trễ) | NFR-03 | 3 | Image + biên dịch và chạy | T2-01 |
| T2-04 | Run Service: lấy samples, chạy, ghi kết quả Redis TTL | FR-06.1–06.3 | 2 | `/problems/{slug}/run`, `/runs/{id}` | T2-01, T3-05 |
| T2-05 | Rate limit Run, kiểm tra khóa cấp, giới hạn 64 KB | FR-06.4–06.7 | 2 | Trả 429/403/413 đúng | T2-04, T3-08 |
| T2-06 | Ghép `wrapper_code` cho bài `function`; chạy bài `stdin_stdout` | FR-05.5 | 2 | Chạy được cả hai chế độ | T2-01, T3-06 |
| T2-07 | Judge API: nhận Submit, tạo job Redis, đẩy hàng đợi | FR-07.2 | 3 | `/problems/{slug}/submissions`, `/submissions/jobs/{id}` | T2-01 |
| T2-08 | Judge Worker: chạy test ẩn, so sánh output, phân loại 7 verdict | FR-07.1, FR-07.3, FR-07.7 | 3 | Verdict đúng cho bộ test mẫu | T2-07, T3-05 |
| T2-09 | Ghi AC idempotent: SHA-256, `ON CONFLICT`, transaction cùng `user_problem_progress` | FR-07.4–07.6, NFR-07 | 3 | Nộp lại cùng mã không tăng `accepted_count` | T2-08, T3-03 |
| T2-10 | Tính `level_up` sau AC mới (gọi `common-levels`) | FR-08.5 | 3 | Trả `level_up` trong kết quả job | T2-09, T3-09 |
| T2-11 | Rate limit Submit, chặn bài chưa mở khóa/chưa publish | FR-07.8–07.9 | 3 | 429/403 đúng | T2-07 |
| T2-12 | Xử lý IE: lỗi hạ tầng không ghi DB, cho nộp lại | FR-07.11 | 3 | Test mô phỏng lỗi worker | T2-08 |
| T2-13 | Ghi `submission_test_results` (tùy chọn) | FR-07.10 | 4 | Có dữ liệu từng test cho AC | T2-09 |
| T2-14 | Kiểm thử bảo mật sandbox (fork bomb, đọc file hệ thống, mạng, output quá lớn) | NFR-03 | 4–5 | Báo cáo kiểm thử, lỗi đã sửa | T2-08 |
| T2-15 | Tổ chức Daily, Sprint Planning, Retrospective; gỡ vướng | | 0–5 | Biên bản họp | |
| T2-16 | Xác nhận view `leaderboard` và các truy vấn xếp hạng theo kỳ/gói bằng dữ liệu mẫu | FR-10.1, FR-10.3 | 3 | Truy vấn đúng, có test đồng hạng | T3-01 |
| T2-17 | API xếp hạng: `/leaderboard`, `/leaderboard/me` | FR-10.1–10.4, FR-10.6 | 4 | Hai endpoint chạy được | T2-16 |
| T2-18 | Cache Redis xếp hạng và tên user; ghép username từ Auth | FR-10.2, FR-10.5 | 4 | Cache hoạt động, có fallback username | T2-17, T1-06 |
| T2-19 | Kiểm thử tích hợp end-to-end: đăng nhập → Run → Submit → Level up → Leaderboard | | 3–5 | Kịch bản chạy lặp lại được, báo cáo | T2-09, T2-17 |
| T2-20 | Tích hợp worker → `ai_service` sau Accepted, idempotency và timeout | FR-13.1–13.3 | 4–5 | AI job chỉ phát sinh sau Accepted; lỗi AI không đổi verdict/rank | T2-09, T3-17 |

### TV3 — Hiếu: Core API (nội dung bài tập)

| Mã | Nhiệm vụ | FR | Sprint | Đầu ra | Phụ thuộc |
|---|---|---|---|---|---|
| T3-01 | Migration/schema đích: Package, Level, progress, leaderboard, forum, AI; loại repo/GitHub khỏi model mới | NFR-08, FR-13 | 0–1 | DB migration có rollback/seed kiểm tra được | |
| T3-02 | Package `common-db`: model/query Go dùng pgx cho các bảng mới | NFR-08 | 1–2 | Package dùng chung cho Core, Run, Submission, Forum, AI | T3-01 |
| T3-03 | Model/CRUD nội bộ cho `user_problem_progress`, `submissions` | FR-07.4–07.6 | 2 | Hàm repository dùng bởi Judge | T3-02 |
| T3-04 | API ngôn ngữ (public và theo bài) | FR-03.1–03.2 | 2 | `/languages`, `/problems/{slug}/languages` | T3-02 |
| T3-05 | API gói bài, danh sách/lọc/chi tiết bài; không bao giờ trả test ẩn | FR-04.1–04.5 | 2 | `/categories`, `/problems`, `/problems/{slug}` | T3-02, T1-04 |
| T3-06 | API template theo ngôn ngữ (không trả `wrapper_code`) | FR-05.1, FR-05.6 | 2 | `/problems/{slug}/template` | T3-02 |
| T3-07 | Admin API: CRUD gói bài, bài, tag; vòng đời draft/published/archived | FR-12.2–12.4, FR-12.8 | 2 | `/admin/categories`, `/admin/problems` | T3-02, T1-04 |
| T3-08 | Package `common-levels`: `get_level_status`, `is_level_unlocked` | FR-08.1–08.4 | 2 | Package + unit test các tình huống mở khóa | T3-02 |
| T3-09 | API `/categories/{slug}/levels` và chặn 403 `LEVEL_LOCKED` ở chi tiết bài | FR-04.6, FR-08.6 | 3 | Trả đúng 3 cấp và trạng thái | T3-08 |
| T3-10 | Admin API: samples, test ẩn, khóa sửa sau publish, ngôn ngữ, template | FR-12.5–12.7 | 2–3 | `/admin/problems/{id}/samples`, `/test-cases`, ... | T3-07 |
| T3-11 | Điều kiện publish và nhân bản bài (clone) | FR-12.4, FR-12.11 | 3 | `/publish`, `/clone` | T3-10 |
| T3-12 | API lịch sử submission, chi tiết và kiểm quyền | FR-09.1–09.5 | 3 | `/me/submissions`, `/submissions/{id}` | T3-03 |
| T3-13 | API landing, thử thách hằng ngày | FR-02.3, FR-04.7 | 4 | `/public/landing`, `/daily-challenge` | T1-06 |
| T3-14 | Admin API: xem submission, thống kê, daily challenge | FR-12.10, FR-12.12–12.13 | 4 | `/admin/submissions`, `/admin/stats` | T3-12 |
| T3-15 | Seed dữ liệu mẫu: tối thiểu 3 gói × 3 cấp × vài bài, kèm test (cả team cùng soạn bài) | | 2–5 | Script seed chạy lặp lại được | T3-10 |
| T3-16 | Kiểm thử bảo mật web: phân quyền, lộ test ẩn, XSS, rate limit | NFR-04, NFR-05 | 5 | Báo cáo bảo mật, lỗi đã sửa | T1-15, T2-11 |
| T3-17 | Repository/schema cho `ai_reviews`, trạng thái và owner check | FR-13.1–13.5 | 3–4 | Migration + query đọc review; không lưu AI score vào leaderboard | T3-01, T3-02 |

### TV4 — Huy: Frontend

| Mã | Nhiệm vụ | FR | Sprint | Đầu ra | Phụ thuộc |
|---|---|---|---|---|---|
| T4-01 | Khởi tạo dự án Next.js + TypeScript, routing, layout, theme, Axios + interceptor refresh token, bộ component dùng chung | NFR-10 | 0–1 | Skeleton chạy được; component dùng chung cho cả TV1 | |
| T4-02 | Landing page | FR-02.1–02.4 | 1 | Trang Landing responsive | |
| T4-03 | Đăng ký, đăng nhập, điều hướng theo vai trò | FR-01 | 1 | Trang auth + route guard | T1-02 |
| T4-04 | Danh sách gói bài, trạng thái khóa, tiến độ cấp | FR-04.1–04.2, FR-08.6 | 2–3 | Trang gói bài | T3-05, T3-09 |
| T4-05 | Danh sách/lọc/tìm bài; trang chi tiết bài | FR-04.3–04.5 | 2 | Trang bài | T3-05 |
| T4-06 | Trang làm bài: Monaco, chọn ngôn ngữ, nạp starter code, Reset, nháp localStorage | FR-03, FR-05 | 2 | Editor hoạt động | T3-04, T3-06 |
| T4-07 | Nút Run + panel kết quả (polling) | FR-06 | 2 | Hiển thị từng sample | T2-04 |
| T4-08 | Nút Submit + panel verdict (polling), hiệu ứng lên cấp | FR-07, FR-08.5 | 3 | Hiển thị verdict, level up; không chờ AI để xác nhận Accepted | T2-07 |
| T4-09 | Lịch sử nộp bài, xem lại code, "Dùng lại code này" | FR-09 | 3 | Trang lịch sử | T3-12 |
| T4-10 | Bảng xếp hạng (lọc kỳ/gói, hạng của tôi) | FR-10 | 4 | Trang xếp hạng | T2-17 |
| T4-11 | Giao diện Admin: user, gói/bài, test, template, kiểm duyệt diễn đàn, thống kê | FR-12 | 3–5 | Bộ trang Admin | T3-07, T3-10, T1-07, T1-16 |
| T4-12 | Responsive, trạng thái loading/lỗi, kiểm tra trên 4 trình duyệt | NFR-10, NFR-12 | 5 | Báo cáo kiểm tra | |
| T4-13 | UI đọc AI review theo trạng thái queued/running/completed/failed | FR-13.2, FR-13.4 | 5 | Trang submission hiển thị score/findings khi API đích có | T3-17, T2-20 |

*Trang Diễn đàn (T1-19) và Hồ sơ (T1-20) do TV1 làm, dùng bộ component và layout của T4-01.*

---

## 4. PHỤ THUỘC GIỮA CÁC THÀNH VIÊN (HỢP ĐỒNG CẦN CHỐT SỚM)

| Bên cần | Cần gì | Từ ai | Hạn chốt |
|---|---|---|---|
| Mọi service | Package `common-auth` (xác thực JWT, `require_admin`) | TV1 | Cuối Sprint 1 |
| Mọi service | Package `common-errors` | TV1 | Sprint 0 |
| Core, Run, Submission, Forum, AI | Package `common-db` (model/query Go + pgx) | TV3 | 08/11 |
| Core, Run, Judge | Package `common-levels` | TV3 | Giữa Sprint 2 |
| Leaderboard, Landing | API `/internal/users` | TV1 | Sprint 2 |
| Judge Worker | Hàm repository ghi `submissions`, `user_problem_progress` | TV3 | Đầu Sprint 3 |
| Run, Judge | Cấu trúc `wrapper_code`, `execute_code` cho bài `function` | TV2 chốt, TV3 nhập dữ liệu | Sprint 2 |
| Giao diện Diễn đàn, Hồ sơ (TV1) | Bộ component và layout dùng chung | TV4 (T4-01) | Cuối Sprint 1 |
| Frontend | OpenAPI/mock cho Run, Submit, Leaderboard, Forum | TV2, TV3, TV1 | Sprint 0 (bản khung), cập nhật mỗi sprint |
| Submission Worker, AI, Frontend | Contract `ai_reviews` và owner/status semantics | TV2, TV3, TV4 | 15/11 |

**Quy tắc đổi hợp đồng:** thay đổi request/response của API đã chốt phải sửa OpenAPI, ghi vào changelog SRS và báo kênh chung trước khi merge.

---

## 5. LỘ TRÌNH SPRINT VÀ MỐC THỰC THI

Các nhãn Sprint dưới đây chỉ để truy vết user story. Lịch có hiệu lực là:

| Mốc | Phạm vi bắt buộc | Owner chính |
|---|---|---|
| 06–07/11 | schema Package/Level, auth, catalog, contract Run/Submit | TV1, TV2, TV3, TV4 |
| 08–11/11 | Run/editor, migration nội dung, admin cơ bản | TV2, TV3, TV4 |
| 12–15/11 | Submit/worker/verdict, progress/level, history và AI schema | TV2, TV3, TV4 |
| 16–19/11 | leaderboard, Forum, AI integration/UI, đóng tính năng 19/11 21:00 | TV1, TV2, TV3, TV4 |
| 20–21/11 | E2E, security, contract regression, tài liệu và sửa lỗi | Cả team |
| 22/11 | Nộp/demo, không thêm tính năng | Cả team |

Repository/GitHub không có task trong lịch này; mọi task liên quan trong bản cũ
được xem là lịch sử legacy.

| Sprint | Thời gian | Mục tiêu | Nhiệm vụ chính | Kết quả demo |
|---|---|---|---|---|
| **0** | 06–07/11 | Chuẩn bị | Chốt SRS, T3-01, T1-01, T1-05, T4-01, Docker Compose, CI, quy ước code | Workspace chạy `docker compose up`, DB dựng từ migration |
| **1** | 06–07/11 | Nền tảng | T1-02–T1-04, T4-01–T4-03, T2-01, T3-02, T1-17 | Đăng ký, đăng nhập thật; Landing; sandbox chạy một file Python |
| **2** | 08–11/11 | Lõi luyện tập | T3-03–T3-08, T3-10, T3-15, T2-02, T2-04–T2-06, T4-04–T4-07, T1-06–T1-07, T1-17 | Chọn gói, chọn bài, mở editor, bấm Run (Python và C++) |
| **3** | 12–15/11 | Chấm và tiến độ | T2-03, T2-07–T2-12, T3-09, T3-11–T3-12, T3-17, T4-08–T4-09 | Submit/verdict/progress/level/history; AI schema sẵn sàng |
| **4** | 16–19/11 | Cộng đồng và AI | T2-17–T2-20, T1-13–T1-16, T1-19–T1-20, T4-10–T4-11, T4-13 | Leaderboard, Forum, AI sau Accepted; đóng tính năng |
| **5** | 20–22/11 | Hoàn thiện | T2-19, T2-14, T3-16, T4-12, T3-15, T1-18, T1-12, sửa lỗi/tài liệu | Demo, regression, nộp dự án |

**Tiêu chí kết thúc mỗi sprint:** các nhiệm vụ trong cột "Kết quả demo" chạy được trên môi trường chung (Docker Compose) và được PO nghiệm thu.

---

## 6. QUY TẮC LÀM VIỆC

### 6.1 Cấu trúc repo đề xuất

```
coding-judge/
├── services/
│   ├── auth/        (TV1)
│   ├── core/        (TV3: nội dung bài; TV1: forum; TV2: leaderboard)
│   ├── run/         (TV2)
│   └── judge/       (TV2)
├── packages/
│   ├── common-auth/     (TV1)
│   ├── common-errors/   (TV1)
│   ├── common-db/       (TV3)
│   └── common-levels/   (TV3)
├── frontend/        (TV4; app/forum và app/profile do TV1)
├── sandbox/         (TV2: Dockerfile từng ngôn ngữ)
├── infra/           (docker-compose, CI)
└── docs/            (SRS, phân công, backlog, test plan)
```

Trong `services/core/`, mỗi người làm trong module riêng (`problems/`, `forum/`, `leaderboard/`) để hạn chế xung đột merge.

### 6.2 Quy trình Git

- Nhánh: `main` (ổn định, demo), `develop` (tích hợp), `feature/<mã-nhiệm-vụ>-<mô-tả-ngắn>` (ví dụ `feature/T2-09-ac-idempotent`).
- Commit theo Conventional Commits: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`; kèm mã nhiệm vụ trong tiêu đề hoặc thân commit.
- Mọi thay đổi vào `develop` qua Pull Request; PR phải có ít nhất 1 reviewer khác người tạo; CI xanh mới merge.
- Review chéo (team 4 người nên xoay vòng): TV1 review TV3; TV3 review TV2; TV2 review TV4 (phần gọi API Judge/Leaderboard) và TV1; TV4 review giao diện của TV1. Thay đổi `common-*` cần ít nhất một người dùng package đó review.

### 6.3 Định nghĩa Hoàn thành (Definition of Done)

- Code đã merge qua Pull Request, ít nhất 1 người review.
- Có unit test cho logic chính và pass CI.
- Đáp ứng tiêu chí chấp nhận trong SRS và được PO nghiệm thu.
- API có mô tả trong Swagger/OpenAPI; không còn lỗi mức nghiêm trọng.
- Không có secret (mật khẩu, khóa) trong repo.

### 6.4 Nghi thức Scrum

| Nghi thức | Tần suất | Chủ trì | Nội dung |
|---|---|---|---|
| Sprint Planning | Đầu sprint | TV2 (SM) cùng TV1 (PO) | Chọn nhiệm vụ từ bảng mục 3 theo mục tiêu sprint |
| Daily | Hằng ngày, tối đa 15 phút | TV2 | Hôm qua, hôm nay, vướng mắc |
| Sprint Review | Cuối sprint | TV1 | Demo, nghiệm thu |
| Retrospective | Cuối sprint | TV2 | Cái tốt, cái cần cải thiện, hành động |

### 6.5 Kiểm thử khi không có QA riêng

- Người viết chức năng viết unit test và tự kiểm tra tiêu chí chấp nhận.
- TV1 (PO) viết test plan và test case (T1-17), đồng thời nghiệm thu từng chức năng.
- TV2 chịu trách nhiệm kiểm thử end-to-end (T2-19) và bảo mật sandbox (T2-14); TV3 chịu trách nhiệm kiểm thử bảo mật web (T3-16).
- Báo lỗi trên issue tracker của repo theo mẫu trong test plan; lỗi nghiêm trọng xử lý trong sprint hiện tại.

---

## 7. CÂN BẰNG TẢI VÀ RỦI RO PHÂN CÔNG

| Vấn đề | Biểu hiện | Đề xuất |
|---|---|---|
| TV2 gánh phần khó nhất (sandbox, Judge) kiêm SM | Dồn việc ở Sprint 3 | Bắt đầu T2-01 từ Sprint 1; dời T2-03 (Java) sang Sprint 4 nếu trễ; cân nhắc giao SM cho TV1 hoặc TV3 |
| TV4 làm gần hết giao diện | Sprint 3–4 nhiều trang cùng lúc | Dùng thư viện UI sẵn có (ví dụ Ant Design, MUI); TV1 đã nhận trang Diễn đàn và Hồ sơ |
| TV1 vừa PO vừa Auth, Forum, test plan | Sprint 4 dồn Forum + UI + Auth phụ trợ | Auth làm xong chủ yếu ở Sprint 1–2; T1-09 và T1-10 làm ở Sprint 3 nếu còn sức; cắt FR-11.6, FR-11.7 (mức Should) trước |
| TV3 nhiều nhiệm vụ nền tảng | Chặn các thành viên khác nếu trễ | Ưu tiên T3-01, T3-02, T3-08 trước; phát hành package sớm dù chưa hoàn chỉnh |
| Ít người hơn bản 5 người | Tổng khối lượng không đổi | Cắt nhiệm vụ mức Could trước (bên dưới); nếu vẫn trễ ở Sprint 3, đề xuất kéo dài thêm 1 sprint thay vì giảm chất lượng Judge |
| Soạn bài mẫu (≥ 30 bài) | Dồn cuối dự án | Mỗi người soạn tối thiểu 6–8 bài từ Sprint 2, TV3 nhập qua seed |

**Cắt đầu tiên khi trễ (thứ tự):** T2-13 (`submission_test_results`), phần daily challenge của T3-13, phần thống kê và daily challenge của T3-14, FR-09.6, T1-10 (giới hạn đăng nhập sai), T1-09 (quên mật khẩu).

---

## 8. THEO DÕI TIẾN ĐỘ

Mỗi thành viên cập nhật cột Trạng thái của nhiệm vụ mình phụ trách trong tài liệu theo dõi chung (bảng tính hoặc issue tracker của repo). Giá trị hợp lệ: `Chưa bắt đầu`, `Đang làm`, `Chờ review`, `Hoàn thành`, `Bị chặn (ghi rõ chặn bởi mã nhiệm vụ nào)`.

## 9. Phối hợp cập nhật v2.0

Stack bám code hiện tại: Next.js/TypeScript phía frontend; Go `net/http`, pgx, Redis list queue và Piston phía backend. Giữ mã task/owner nhưng không tạo repo/GitHub flow. T4-11 triển khai Admin từ 08/11; T2-20/T3-17/T4-13 là AI review mới. Lịch có hiệu lực 06–22/11/2026; quy tắc tích hợp xem `04_KE_HOACH_TICH_HOP.md`.
