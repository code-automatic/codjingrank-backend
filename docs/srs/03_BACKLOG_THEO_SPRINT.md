# BACKLOG USER STORY THEO SPRINT

**Thành viên:** TV1 = Quốc Thịnh; TV2 = Tùng (Hiếu hỗ trợ); TV3 = Hiếu; TV4 = Huy (Quốc Thịnh hỗ trợ). Giữ nguyên mã story/task và owner chính.

## Dự án: Coding Judge Platform

| | |
|---|---|
| **Phiên bản** | 2.0 (khớp SRS 2.0 và bảng phân công 2.0, team 4 người) |
| **Ngày cập nhật** | 05/11/2026 |
| **Trạng thái** | Backlog đích; chưa phải trạng thái đã triển khai |
| **Tài liệu gốc** | `01_SRS_Coding_Judge_Platform.md`, `02_PHAN_CONG_NHIEM_VU.md` |

**Cách đọc:** mỗi user story có mã `US-xx`, trỏ về yêu cầu `FR-xx.y` trong SRS và nhiệm vụ `Tn-xx` trong bảng phân công. Cột **Tiêu chí chấp nhận** là điều kiện để PO nghiệm thu. Ưu tiên theo SRS: **M** (Must), **S** (Should), **C** (Could).

---

## SPRINT 0 (Tuần 1): Chuẩn bị

**Mục tiêu:** Mọi người cùng chạy được hệ thống khung, thống nhất hợp đồng API.

| Mã | User story | FR | Nhiệm vụ | Ưu tiên | Tiêu chí chấp nhận |
|---|---|---|---|---|---|
| US-01 | Là lập trình viên, tôi muốn dựng DB bằng một lệnh từ schema để mọi người dùng cùng một cấu trúc dữ liệu | NFR-08 | T3-01 | M | `docker compose run --rm migrate-judge up` tạo đủ bảng, view `leaderboard` và ràng buộc của `coding_judge_schema.sql` |
| US-02 | Là lập trình viên, tôi muốn có định dạng lỗi chung để FE xử lý lỗi thống nhất | SRS 2.3 | T1-05 | M | Mọi service trả lỗi theo `{error:{code,message,details}}` |
| US-03 | Là lập trình viên, tôi muốn có khung Frontend và Axios đã cấu hình token để bắt đầu làm giao diện | NFR-10 | T4-01 | M | `npm run dev` chạy; interceptor tự refresh token một lần khi 401 |
| US-04 | Là cả team, chúng tôi muốn chạy toàn bộ hệ thống bằng Docker Compose và CI để tích hợp liên tục | NFR-11 | (Git Lead, TV2) | M | `docker compose up` khởi động Postgres, Redis, các service; CI chạy lint và test |
| US-05 | Là PO, tôi muốn chốt các điểm ở SRS mục 8 để tránh sửa thiết kế giữa chừng | | T1-11 | M | Cột "Team chốt" của SRS mục 8 được điền đủ |

---

## SPRINT 1 (Tuần 2–3): Nền tảng

**Mục tiêu:** Người dùng đăng ký, đăng nhập thật; Landing hoạt động; sandbox chạy được Python.

| Mã | User story | FR | Nhiệm vụ | Ưu tiên | Tiêu chí chấp nhận |
|---|---|---|---|---|---|
| US-06 | Là khách, tôi muốn đăng ký tài khoản bằng email, username, mật khẩu để bắt đầu luyện tập | FR-01.1, FR-01.2 | T1-02, T4-03 | M | Đăng ký trùng email hoặc username trả 409 với thông báo rõ; mật khẩu dưới 8 ký tự bị từ chối; mật khẩu lưu dạng hash |
| US-07 | Là người dùng, tôi muốn đăng nhập và duy trì phiên để không phải đăng nhập lại liên tục | FR-01.3, FR-01.4 | T1-02, T1-03, T4-03 | M | Đăng nhập trả access và refresh token; token hết hạn tự refresh một lần; đăng xuất thu hồi refresh token |
| US-08 | Là hệ thống, tôi muốn mọi service xác thực JWT bằng cùng một thư viện để phân quyền nhất quán | FR-01.5 | T1-04 | M | User thường gọi API admin nhận 403; token sai chữ ký nhận 401 |
| US-09 | Là khách, tôi muốn xem trang giới thiệu có nút đăng ký/đăng nhập để hiểu hệ thống làm gì | FR-02.1, FR-02.2, FR-02.4 | T4-02 | M | Nút đổi thành "Vào luyện tập" khi đã đăng nhập; hiển thị tốt trên mobile, tablet, desktop |
| US-10 | Là lập trình viên backend, tôi muốn có model/query Go dùng pgx dùng chung để không viết lặp | NFR-08 | T3-02 | M | Core, Run, Submission, Forum và AI dùng `common-db`; truy vấn mẫu chạy được |
| US-11 | Là Judge, tôi muốn chạy một file Python trong container có giới hạn để nền tảng chấm bài sẵn sàng | NFR-03 | T2-01 | M | Chạy được mã với input; vượt thời gian bị dừng; không truy cập được mạng; không chạy bằng root |

---

## SPRINT 2 (Tuần 4–5): Lõi luyện tập

**Mục tiêu:** Người học chọn gói, chọn bài, mở editor, bấm Run (Python và C++). Admin tạo được nội dung.

| Mã | User story | FR | Nhiệm vụ | Ưu tiên | Tiêu chí chấp nhận |
|---|---|---|---|---|---|
| US-12 | Là Admin, tôi muốn tạo gói bài và bài ở trạng thái nháp để chuẩn bị nội dung | FR-12.2, FR-12.3 | T3-07, T4-11 | M | CRUD gói và bài hoạt động; bài mới có `status = draft` |
| US-13 | Là Admin, tôi muốn thêm sample và test ẩn cho bài để hệ thống có dữ liệu chấm | FR-12.5 | T3-10, T4-11 | M | Thêm, sửa, xóa test khi bài còn draft; có `display_order`, `is_active`, `points` |
| US-14 | Là Admin, tôi muốn gán ngôn ngữ và nhập code mẫu cho từng ngôn ngữ của bài | FR-12.7 | T3-10, T4-11 | M | Không tạo được template cho ngôn ngữ chưa gán vào bài |
| US-15 | Là người học, tôi muốn xem danh sách ngôn ngữ và chọn ngôn ngữ phù hợp với bài | FR-03.1, FR-03.2, FR-03.3 | T3-04, T4-06 | M | Chỉ thấy ngôn ngữ bài hỗ trợ; đổi ngôn ngữ nạp lại starter code tương ứng |
| US-16 | Là người học, tôi muốn xem danh sách gói bài và bài theo cấp độ để chọn bài phù hợp | FR-04.1, FR-04.2, FR-04.3 | T3-05, T4-04, T4-05 | M | Chỉ hiển thị bài `published`; mỗi gói chia ba cấp |
| US-17 | Là người học, tôi muốn lọc và tìm bài theo tên, tag, độ khó, trạng thái đã giải | FR-04.4 | T3-05, T4-05 | S | Bộ lọc kết hợp được; phân trang đúng |
| US-18 | Là người học, tôi muốn xem đề bài đầy đủ trước khi làm | FR-04.5 | T3-05, T4-05 | M | Có mô tả, ràng buộc, sample, giới hạn thời gian và bộ nhớ, tag |
| US-19 | Là người học, tôi muốn soạn code trong editor có highlight và nút Reset | FR-05.1, FR-05.2, FR-05.3 | T3-06, T4-06 | M | Reset đưa về starter code; không lộ `wrapper_code` qua API |
| US-20 | Là người học, tôi muốn nháp code được tự lưu để không mất khi tải lại trang | FR-05.4 | T4-06 | S | Tải lại trang vẫn còn code theo (bài, ngôn ngữ) |
| US-21 | Là người học, tôi muốn Run với test mẫu để kiểm tra nhanh mà không ảnh hưởng thành tích | FR-06.1, FR-06.2, FR-06.3 | T2-04, T4-07 | M | Mỗi sample hiện output thực tế, kỳ vọng, thời gian, bộ nhớ; không có dòng nào được ghi vào PostgreSQL |
| US-22 | Là hệ thống, tôi muốn giới hạn Run để tránh lạm dụng | FR-06.4, FR-06.6 | T2-05 | M | Vượt 10 lần/phút trả 429 kèm `Retry-After`; mã quá 64 KB trả 413 |
| US-23 | Là hệ thống, tôi muốn chặn Run khi bài thuộc cấp chưa mở | FR-06.5, FR-08.4 | T2-05, T3-08 | M | Trả 403 `LEVEL_LOCKED` khi gọi API trực tiếp |
| US-24 | Là Judge, tôi muốn chạy được bài `function` và bài `stdin_stdout` | FR-05.5 | T2-06 | M | Cả hai chế độ chạy đúng với bài mẫu |
| US-25 | Là Judge, tôi muốn có thêm runtime C++17 | NFR-03 | T2-02 | M | Biên dịch lỗi trả CE; chạy đúng với sample |
| US-26 | Là Admin, tôi muốn xem, khóa, mở khóa người dùng và đổi vai trò | FR-01.6, FR-12.1 | T1-06, T1-07, T4-11 | M | Tài khoản bị khóa không đăng nhập được |
| US-27 | Là PO, tôi muốn có test plan để kiểm thử có hệ thống | | T1-17 | M | Tài liệu test plan và bộ test case cho các chức năng của Sprint 1–2 |

---

## SPRINT 3 (Tuần 6–7): Chấm và tiến độ

**Mục tiêu:** Submit đầy đủ: verdict, lưu AC, lên cấp, lịch sử.

| Mã | User story | FR | Nhiệm vụ | Ưu tiên | Tiêu chí chấp nhận |
|---|---|---|---|---|---|
| US-28 | Là người học, tôi muốn Submit để chạy toàn bộ test và nhận verdict | FR-07.1, FR-07.2, FR-07.3 | T2-07, T2-08, T4-08 | M | Trạng thái Pending → Running → verdict cuối; hiện verdict và số test qua; giữ cơ chế lưu submission hiện tại |
| US-29 | Là người học, tôi muốn thấy đúng loại lỗi (CE, RE, TLE, MLE, WA) để biết sửa gì | FR-07.3 | T2-08 | M | Mỗi loại lỗi được tái hiện bằng bài mẫu và trả đúng verdict |
| US-30 | Là người học, tôi muốn test ẩn không bị lộ đáp án | FR-07.7, NFR-05 | T2-08, T3-05 | M | API người học không bao giờ trả `problem_test_cases`; chỉ hiện pass/fail |
| US-31 | Là người học, tôi muốn bài Accepted được lưu để xem lại sau | FR-07.4, FR-07.5 | T2-09 | M | AC mới tạo một dòng `submissions` và upsert `user_problem_progress` trong một transaction; lỗi giữa chừng rollback cả hai |
| US-32 | Là người học, tôi muốn nộp lại cùng một mã mà không làm sai thống kê | FR-07.6, BR-06 | T2-09 | M | Nộp lại mã giống hệt trả AC với `duplicate = true`; `accepted_count` không tăng |
| US-33 | Là hệ thống, tôi muốn giới hạn tần suất nộp bài | FR-07.8 | T2-11 | M | Nộp nhanh hơn 1 lần/5 giây trả 429 |
| US-34 | Là hệ thống, tôi muốn chặn Submit khi bài chưa publish hoặc cấp chưa mở | FR-07.9 | T2-11 | M | Trả 403 đúng mã lỗi |
| US-35 | Là người học, tôi muốn được thử lại khi hệ thống chấm gặp lỗi nội bộ | FR-07.11 | T2-12 | M | IE không ghi DB; nộp lại được ngay |
| US-36 | Là người học, tôi muốn Cấp 2 tự mở khi tôi giải hết Cấp 1 của gói | FR-08.1, FR-08.2, FR-08.3 | T3-08, T3-09, T4-04 | M | Giải bài cuối của Cấp 1 thì `/levels` báo Cấp 2 `unlocked`; cấp không có bài không chặn cấp sau |
| US-37 | Là người học, tôi muốn thấy hiệu ứng khi lên cấp | FR-08.5 | T2-10, T4-08 | S | Kết quả Submit có `level_up` đúng gói và đúng cấp; FE hiển thị thông báo |
| US-38 | Là người học, tôi muốn thấy bài thuộc cấp khóa có biểu tượng khóa | FR-04.6, FR-08.6 | T3-09, T4-04 | M | Mở bài bị khóa trực tiếp bằng URL trả 403 |
| US-39 | Là người học, tôi muốn xem lịch sử các lần nộp Accepted | FR-09.1, FR-09.2 | T3-12, T4-09 | M | Lọc theo bài, ngôn ngữ; phân trang; sắp xếp mới nhất trước |
| US-40 | Là người học, tôi muốn xem lại mã đã nộp và dùng lại | FR-09.3, FR-09.4 | T3-12, T4-09 | M, S | Nút "Dùng lại code này" nạp mã vào editor |
| US-41 | Là người học, tôi muốn chỉ mình tôi xem được mã của tôi | FR-09.5 | T3-12 | M | Truy cập submission của người khác trả 404; Admin xem được tất cả |
| US-42 | Là Judge, tôi muốn có thêm runtime Java 17 | NFR-03 | T2-03 | M | Biên dịch và chạy đúng với bài mẫu |
| US-43 | Là Admin, tôi muốn publish bài chỉ khi đủ điều kiện và nhân bản bài để sửa test an toàn | FR-12.4, FR-12.6, FR-12.11 | T3-10, T3-11 | M, S | Publish thiếu sample/test/template bị từ chối; sửa test của bài đã publish trả 409 `PROBLEM_PUBLISHED_LOCKED`; clone tạo bản nháp |
| US-44 | Là người dùng, tôi muốn sửa hồ sơ và đổi mật khẩu | FR-01.8 | T1-08 | S | Đổi mật khẩu cần mật khẩu cũ |
| US-45 | Là PO, tôi muốn kiểm tra các truy vấn xếp hạng sớm để Sprint 4 không bị chậm | FR-10.1 | T2-16 | M | Truy vấn đúng với dữ liệu mẫu có đồng hạng |

---

## SPRINT 4 (Tuần 8–9): Cộng đồng

**Mục tiêu:** Bảng xếp hạng, diễn đàn, Admin hoàn thiện.

| Mã | User story | FR | Nhiệm vụ | Ưu tiên | Tiêu chí chấp nhận |
|---|---|---|---|---|---|
| US-46 | Là người học, tôi muốn xem bảng xếp hạng theo số bài đã giải | FR-10.1, FR-10.2 | T2-17, T4-10 | M | Đồng số bài thì xét số lần Accepted; đồng số liệu thì đồng hạng |
| US-47 | Là khách, tôi muốn xem top bảng xếp hạng công khai | FR-10.6 | T2-17 | M | Guest xem được scope `all` (top 50) |
| US-48 | Là người học, tôi muốn lọc xếp hạng theo tuần, tháng, gói bài và xem hạng của mình | FR-10.3, FR-10.4 | T2-17, T4-10 | S | `/leaderboard/me` trả đúng hạng khi ngoài top |
| US-49 | Là hệ thống, tôi muốn bảng xếp hạng nhanh và có tên người dùng | FR-10.5 | T2-18 | S | Cache 60 giây; khi Auth lỗi hiện username dự phòng |
| US-50 | Là người học, tôi muốn đăng bài viết thảo luận, hỏi đáp, chia sẻ lời giải | FR-11.1, FR-11.2 | T1-13, T1-19 | M | Hỗ trợ Markdown và code; gắn được với bài tập và ngôn ngữ |
| US-51 | Là người học, tôi muốn sửa và xóa bài viết của mình | FR-11.3 | T1-13 | M | Xóa mềm; chỉ chủ bài sửa/xóa được |
| US-52 | Là người học, tôi muốn bình luận và trả lời bình luận | FR-11.4 | T1-14, T1-19 | M | Chỉ lồng một cấp; sai cấp trả 422 |
| US-53 | Là người học, tôi muốn thích bài viết | FR-11.5 | T1-14 | M | Mỗi user chỉ thích một lần mỗi bài; bỏ thích được |
| US-54 | Là người học, tôi muốn tìm và sắp xếp bài viết | FR-11.6, FR-11.7 | T1-13, T1-14 | S | Sắp xếp theo mới nhất, nhiều like, nhiều lượt xem |
| US-55 | Là người học, tôi muốn không bị lộ lời giải khi chưa giải bài | FR-11.8 | T1-15 | S | Đăng `solution` cho bài chưa giải trả 403 `SOLUTION_NOT_ALLOWED` |
| US-56 | Là người dùng, tôi muốn nội dung diễn đàn an toàn | FR-11.10 | T1-15 | M | Script chèn trong Markdown không thực thi (có test XSS) |
| US-57 | Là Admin, tôi muốn ẩn hoặc xóa bài viết và bình luận vi phạm | FR-12.9 | T1-16, T4-11 | M | Nội dung bị ẩn không hiện cho người học |
| US-58 | Là Admin, tôi muốn xem toàn bộ submission và thống kê | FR-12.10, FR-12.12 | T3-14, T4-11 | S, C | Lọc theo user, bài; thống kê không có tỉ lệ AC |
| US-59 | Là khách, tôi muốn Landing có thống kê công khai | FR-02.3 | T3-13, T4-02 | S | Hiển thị số bài, số người dùng, lượt giải, top 5 |
| US-60 | Là người học, tôi muốn xem thử thách hằng ngày | FR-04.7, FR-12.13 | T3-13, T3-14 | C | Hiển thị bài của ngày; Admin đặt được bài theo ngày |
| US-61 | Là người dùng, tôi muốn đặt lại mật khẩu khi quên | FR-01.7 | T1-09 | S | Token một lần, hết hạn sau 30 phút |
| US-62 | Là hệ thống, tôi muốn hạn chế dò mật khẩu | FR-01.10 | T1-10 | S | Quá 5 lần sai/phút/IP bị chặn tạm thời |
| US-63 | Là người dùng, tôi muốn có trang hồ sơ | FR-01.8 | T1-20 | S | Hiển thị và sửa thông tin cá nhân |
| US-64 | Là người học, tôi muốn xem chi tiết từng test của bài đã giải | FR-07.10, FR-09.6 | T2-13 | C | Có dữ liệu `submission_test_results` cho AC |

---

## SPRINT 5 (Tuần 10–11): Hoàn thiện

**Mục tiêu:** Kiểm thử toàn diện, bảo mật, tối ưu, tài liệu, demo.

| Mã | User story | FR | Nhiệm vụ | Ưu tiên | Tiêu chí chấp nhận |
|---|---|---|---|---|---|
| US-65 | Là thành viên phụ trách kiểm thử, tôi muốn có kịch bản end-to-end để kiểm tra luồng chính | | T2-19 | M | Kịch bản Đăng nhập → Run → Submit → Level up → Leaderboard chạy lặp lại được và pass |
| US-66 | Là team, chúng tôi muốn sandbox chống được các kiểu tấn công thông thường | NFR-03 | T2-14 | M | Fork bomb, đọc file hệ thống, truy cập mạng, output quá lớn đều bị chặn và không làm sập worker |
| US-67 | Là team, chúng tôi muốn kiểm tra phân quyền và lộ dữ liệu | NFR-04, NFR-05 | T3-16 | M | Không có API người học nào lộ test ẩn; user thường không gọi được `/admin/*` |
| US-68 | Là người dùng, tôi muốn giao diện chạy tốt trên các trình duyệt phổ biến và thiết bị | NFR-10, NFR-12 | T4-12 | M | Kiểm tra Chrome, Edge, Firefox, Safari; mobile, tablet, desktop |
| US-69 | Là người học, tôi muốn hệ thống có đủ bài để luyện tập | | T3-15, mọi thành viên | M | Tối thiểu 30 bài trải đều 3 cấp và 3 ngôn ngữ; mỗi bài có sample, test ẩn, template |
| US-70 | Là người dùng mới, tôi muốn có hướng dẫn sử dụng | | T1-18 | S | Tài liệu hướng dẫn và báo cáo kiểm thử hoàn chỉnh |
| US-71 | Là team, chúng tôi muốn đo hiệu năng chấm | NFR-01, NFR-02 | T2-19, T2-14 | S | 20 submission đồng thời; bài chuẩn có kết quả dưới 10 giây |
| US-72 | Là PO, tôi muốn nghiệm thu cuối cùng và chốt bản demo | | T1-11, T1-12 | M | Danh sách FR mức Must đều đạt; SRS cập nhật phiên bản cuối |

## BỔ SUNG V2 — FORUM VÀ AI REVIEW

Các story dưới đây bổ sung cho backlog nguồn. AI review chưa implement và chỉ
được coi là hoàn thành khi có migration, service-to-service contract và test.

| Mã | User story | FR | Nhiệm vụ | Ưu tiên | Tiêu chí chấp nhận |
|---|---|---|---|---|---|
| US-73 | Là hệ thống, tôi muốn tự tạo AI review sau Accepted | FR-13.1–13.3 | T2-20, T3-17 | M | Accepted tạo tối đa một review; Run/WA/CE không tạo; AI lỗi không đổi verdict/rank |
| US-74 | Là người học, tôi muốn xem điểm và nhận xét AI của bài Accepted | FR-13.2, FR-13.4 | T4-13 | M | Owner xem được queued/running/completed/failed; user khác nhận 404/403 |
| US-75 | Là team, chúng tôi muốn diễn đàn độc lập với judge và repo | FR-11.1–11.10 | T1-13–T1-16, T1-19 | M | Post/comment/like/moderation chạy qua Forum service; không gọi GitHub/repository |
| US-76 | Là frontend, tôi muốn contract phân biệt API live và API đích | NFR-08 | T1-12, T4-13 | M | Client không gọi `/api/v1` AI/Forum/Leaderboard trước khi backend có contract test |

---

## TỔNG KẾT PHẠM VI THEO ƯU TIÊN

| Ưu tiên | Ý nghĩa | Cách xử lý khi trễ |
|---|---|---|
| **M** (Must) | Bắt buộc có trong bản demo | Không cắt; huy động thành viên khác hỗ trợ |
| **S** (Should) | Nên có | Làm sau khi xong Must của sprint |
| **C** (Could) | Có thể bỏ | Cắt đầu tiên: US-58 (phần thống kê), US-60, US-64 |

## Lịch thực thi v2.0

Các Sprint 0–5 chỉ giữ nhóm story để truy vết. Lịch có hiệu lực: 06–07/11
nền tảng/schema; 08–11/11 Run/editor/admin; 12–15/11 Submit/progress/AI schema;
16–19/11 leaderboard/Forum/AI integration và đóng tính năng; 20–21/11
kiểm thử/sửa lỗi; 22/11 nộp. Không có story tạo repository/GitHub. Xem
`04_KE_HOACH_TICH_HOP.md`; backend bám Go `net/http`, pgx và Redis list queue.
