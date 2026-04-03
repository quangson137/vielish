# Domain Glossary

Danh sách thuật ngữ nghiệp vụ chuẩn dùng xuyên suốt tài liệu và giao tiếp trong dự án. Khi có xung đột giữa cách diễn đạt ở các nơi khác nhau, ưu tiên định nghĩa trong file này.

---

## B

**Bài nghe (Listening Lesson)**
Một bài luyện kỹ năng nghe hoàn chỉnh — gồm file âm thanh gốc, lời thoại song ngữ (Anh–Việt), và các câu hỏi kiểm tra hiểu bài. Được phân theo trình độ để phù hợp với từng nhóm người học.
*(Planned — chưa triển khai trong MVP hiện tại)*

---

## C

**Câu hỏi bài nghe (Listening Question)**
Câu hỏi gắn với một bài nghe nhằm kiểm tra mức độ hiểu của người học. Ba dạng câu hỏi: điền từ vào chỗ trống, trắc nghiệm một đáp án, đúng/sai.
*(Planned)*

**Chế độ học (Learn Mode)**
Hình thức tiếp xúc từ vựng lần đầu — người học xem từng flashcard và tự đánh giá mức độ nhớ sau khi xem mặt sau. Kết thúc khi đã xem hết toàn bộ từ trong chủ đề.

**Chế độ kiểm tra (Quiz)**
Hình thức đo lường từ vựng theo chủ đề — người học nhìn từ tiếng Anh và chọn nghĩa tiếng Việt đúng trong bốn đáp án. Câu hỏi được tạo tự động từ từ vựng trong chủ đề; kết quả kiểm tra không được lưu lại.

**Chế độ ôn tập (Review Mode)**
Hình thức ôn lại các từ đến hạn theo lịch SRS — chỉ hiển thị những từ mà người học đã học qua và đã đến thời điểm cần ôn lại. Tối đa 20 từ mỗi phiên, ưu tiên những từ quá hạn nhất.

**Chủ đề (Topic)**
Tập hợp từ vựng cùng một lĩnh vực hoặc ngữ cảnh sử dụng — ví dụ: Thức ăn, Du lịch, Công việc. Mỗi chủ đề gắn với một trình độ và là đơn vị để người học chọn khi bắt đầu học.

---

## Đ

**Đánh giá ôn tập (Review Quality)**
Mức độ tự đánh giá của người học sau mỗi lần xem flashcard. Là tín hiệu duy nhất hệ thống dùng để điều chỉnh lịch ôn tập tiếp theo. Gồm ba mức:

| Mức | Ý nghĩa |
|-----|---------|
| **Khó** | Quên hoặc nhớ rất mờ — cần ôn lại sớm |
| **Ổn** | Nhớ được nhưng còn phải suy nghĩ |
| **Dễ** | Nhớ ngay lập tức, không cần suy nghĩ |

---

## F

**Flashcard (Thẻ học)**
Thẻ học hai mặt cho một từ vựng:
- **Mặt trước** — từ tiếng Anh, phiên âm, từ loại
- **Mặt sau** — nghĩa tiếng Việt, định nghĩa tiếng Anh, câu ví dụ song ngữ

Người học tự lật thẻ khi đã sẵn sàng, rồi tự đánh giá mức độ nhớ.

---

## H

**Hệ số ghi nhớ (Ease Factor)**
Chỉ số phản ánh mức độ dễ hay khó của một từ đối với một người học cụ thể. Hệ số này tự động điều chỉnh sau mỗi lần ôn dựa trên đánh giá — từ được nhớ tốt sẽ có hệ số cao hơn, kéo dài khoảng cách ôn nhanh hơn. Không bao giờ giảm xuống dưới một ngưỡng tối thiểu để đảm bảo người học luôn có cơ hội cải thiện.

---

## K

**Khoảng cách ôn tập (Review Interval)**
Số ngày cho đến lần ôn tiếp theo của một từ. Tăng dần sau mỗi lần ôn thành công — từ 1 ngày, 6 ngày, rồi tăng theo hệ số ghi nhớ. Reset về 1 ngày nếu người học đánh giá là Khó.

---

## L

**Lịch ôn tập (Review Schedule)**
Kế hoạch ôn tập tự động do hệ thống tính toán cho mỗi từ của mỗi người học. Xác định thời điểm tiếp theo cần ôn lại một từ, dựa trên khoảng cách ôn tập hiện tại.

---

## N

**Nghĩa tiếng Việt (Vietnamese Meaning)**
Bản dịch tiếng Việt của một từ tiếng Anh. Đây là nội dung cốt lõi mà người học cần nắm — hiển thị ở mặt sau flashcard và là đáp án đúng trong chế độ kiểm tra.

**Người học (Learner)**
Người sử dụng ứng dụng để học tiếng Anh. Mỗi người học có tiến trình và lịch ôn tập hoàn toàn độc lập, không ảnh hưởng lẫn nhau dù học cùng nội dung.

---

## P

**Phiên đăng nhập (Session)**
Trạng thái xác thực của một người học sau khi đăng nhập thành công. Có thời hạn và có thể được gia hạn hoặc thu hồi chủ động khi đăng xuất.

---

## S

**SM-2 (SuperMemo 2)**
Thuật toán tính lịch ôn tập được Vielish áp dụng. Sau mỗi lần ôn, thuật toán điều chỉnh hệ số ghi nhớ và khoảng cách ôn dựa trên mức đánh giá của người học, nhằm tối ưu thời điểm ôn trước khi từ bị quên.

**SRS (Spaced Repetition System — Hệ thống lặp lại ngắt quãng)**
Phương pháp học tập khoa học: ôn lại nội dung vào đúng thời điểm người học chuẩn bị quên. Giúp ghi nhớ lâu dài với tổng thời gian ôn tập ít hơn so với học thuộc lòng truyền thống.

---

## T

**Tiến trình học từ (Learning Progress)**
Hồ sơ theo dõi hành trình học của một người học với một từ cụ thể. Ghi lại: hệ số ghi nhớ, khoảng cách ôn hiện tại, thời điểm cần ôn tiếp theo, và số lần đã ôn. Chỉ tồn tại sau khi người học chủ động ôn từ đó lần đầu.

**Trình độ (Level)**
Cấp độ tiếng Anh — áp dụng cho người học, chủ đề, và từ vựng. Ba cấp: Cơ bản, Trung cấp, Nâng cao. Trình độ của người học ảnh hưởng đến nội dung được gợi ý, không giới hạn quyền truy cập.

**Từ vựng (Word)**
Đơn vị học tập cơ bản trong ứng dụng — một từ tiếng Anh kèm đầy đủ thông tin: phiên âm IPA, từ loại, nghĩa tiếng Việt, định nghĩa tiếng Anh, câu ví dụ song ngữ, audio phát âm, và hình minh họa.

**Từ đến hạn ôn (Due Words)**
Những từ trong tiến trình học của một người học đã đến hoặc quá thời điểm cần ôn lại. Đây là danh sách xuất hiện trong chế độ ôn tập mỗi ngày.

**Từ loại (Part of Speech)**
Phân loại ngữ pháp của một từ tiếng Anh: danh từ, động từ, tính từ, trạng từ… Giúp người học hiểu cách sử dụng từ trong câu.

---

## X

**Xác thực (Authentication)**
Quá trình xác minh danh tính của người học khi đăng nhập. Người học cung cấp email và mật khẩu; hệ thống cấp phiên đăng nhập có thời hạn.
