# Domain Context

## Problem Space

Người học tiếng Anh tại Việt Nam thường gặp hai rào cản chính:

1. **Từ vựng không đủ bền vững** — Học thuộc lòng theo kiểu truyền thống không có hệ thống lặp lại khoa học nên quên nhanh.
2. **Khả năng nghe kém** — Ít được tiếp xúc với tiếng Anh thực tế, khó theo kịp nhịp độ và âm điệu của người bản ngữ.

Vielish giải quyết hai vấn đề này bằng cách kết hợp **flashcard có lịch ôn tập tự động (SRS)** cho từ vựng và **bài nghe có bài tập tương tác** cho kỹ năng nghe.

---

## Stakeholders

| Vai trò | Mô tả |
|---------|-------|
| **Người học** | Người dùng chính — học viên tiếng Anh ở mọi trình độ (beginner → advanced), nói tiếng Việt |
| **Content admin** | Người tạo và quản lý nội dung học (chủ đề, từ vựng, bài nghe) — hiện tại seed trực tiếp vào DB |

---

## MVP Features

### 1. Vocabulary (Từ vựng)

Người dùng học từ vựng theo **chủ đề** (Travel, Food, Business…). Mỗi chủ đề gồm ~20–50 từ, phân theo trình độ.

**Ba chế độ học:**

| Chế độ | Luồng | Mục tiêu |
|--------|-------|---------|
| **Learn** (Học mới) | Xem flashcard → lật → tự đánh giá Hard/OK/Easy | Giới thiệu từ mới, ghi nhớ lần đầu |
| **Review** (Ôn tập) | SRS nhắc đúng thời điểm → flashcard → đánh giá | Củng cố từ đã học theo thuật toán SM-2 |
| **Quiz** (Kiểm tra) | Chọn nghĩa đúng từ 4 đáp án | Đo lường mức độ ghi nhớ theo chủ đề |

### 2. Listening (Nghe hiểu) — planned

Người dùng nghe audio → trả lời câu hỏi (điền từ / trắc nghiệm / đúng-sai). Nội dung phân cấp theo trình độ, có phụ đề song ngữ.

---

## Key Business Rules

### Từ vựng

- Một từ (`Word`) thuộc đúng một chủ đề (`Topic`).
- Mỗi chủ đề có một trình độ duy nhất: `beginner | intermediate | advanced`.
- Tiến trình học (`UserWordProgress`) là per-user per-word — hai người dùng học cùng từ nhưng có lịch ôn tập độc lập.
- **Lần đầu xem một từ** (trong chế độ Learn) tạo ra bản ghi `UserWordProgress` với `review_count = 0`.
- **Chỉ từ đã có `UserWordProgress`** mới xuất hiện trong Review (không hiển thị từ chưa học).

### SRS (SM-2)

- Sau mỗi lần ôn, người dùng tự đánh giá theo 3 mức: **Hard (1) / OK (3) / Easy (5)**.
- Thuật toán SM-2 tính `ease_factor` (≥ 1.3) và `interval_days` cho lần ôn tiếp theo.
  - Hard → interval reset về 1 ngày, ease_factor giảm.
  - OK → interval tăng theo lịch tiêu chuẩn, ease_factor giảm nhẹ.
  - Easy → interval tăng nhanh, ease_factor tăng.
- `next_review_at` được tính từ thời điểm ôn + `interval_days * 24h`.
- Giới hạn 20 từ hiển thị mỗi phiên Review (ordered by `next_review_at ASC`).

### Quiz

- Câu hỏi được sinh động (không lưu vào DB) từ từ vựng trong chủ đề.
- Mỗi câu: hiển thị từ tiếng Anh → chọn 1/4 nghĩa tiếng Việt.
- 3 đáp án nhiễu được lấy ngẫu nhiên từ các từ khác trong cùng chủ đề.
- Kết quả quiz (score / total / chi tiết) trả về ngay, không lưu lại.

### Auth

- Người dùng đăng ký bằng email + mật khẩu.
- Session gồm: **access token** (JWT, 1h) + **refresh token** (random hex, 7 ngày, lưu Redis).
- Các endpoint `/api/review/*` và `/api/quiz/*` yêu cầu xác thực.
- Các endpoint `/api/topics`, `/api/words/*` là public (không cần đăng nhập).

---

## Key Workflows

### Workflow 1: Học từ mới (Learn)

```
1. Người dùng mở chủ đề → xem danh sách từ
2. Bấm "Học từ mới" → vào chế độ flashcard
3. Xem mặt trước (từ + IPA + từ loại)
4. Nhấn lật → xem mặt sau (nghĩa + ví dụ)
5. Tự đánh giá: Hard / OK / Easy
   → API POST /api/review/:wordId {"quality": 1|3|5}
   → Tạo/cập nhật UserWordProgress
6. Lặp lại đến hết chủ đề → màn hình "Hoàn thành"
```

### Workflow 2: Ôn tập SRS (Review)

```
1. Người dùng vào trang Ôn tập
   → API GET /api/review/due (trả về tối đa 20 từ đến hạn)
2. Nếu không có từ nào → thông báo "Không có từ cần ôn tập"
3. Flashcard từng từ → đánh giá Hard/OK/Easy
   → API POST /api/review/:wordId
   → Cập nhật UserWordProgress (ease_factor, interval, next_review_at)
4. Kết thúc khi hết danh sách
```

### Workflow 3: Làm bài kiểm tra (Quiz)

```
1. Người dùng bấm "Làm bài kiểm tra" trong chủ đề
   → API GET /api/quiz/:topicId (sinh câu hỏi động)
2. Chọn đáp án cho từng câu (không giới hạn thời gian)
3. Bấm "Nộp bài"
   → API POST /api/quiz/:topicId {"answers": [...]}
   → Server chấm điểm, trả về score + chi tiết đúng/sai
4. Xem kết quả, quay lại chủ đề
```

---

## Out of Scope (MVP)

- AI chatbot / luyện hội thoại
- Nhận dạng giọng nói
- Gamification (bảng xếp hạng, huy hiệu)
- Tính năng mạng xã hội
- Ứng dụng mobile
- Quản lý content qua giao diện (admin UI)
