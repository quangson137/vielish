# API Design Best Practices

---

## 1. Design từ góc nhìn Customer (Consumer-first)

### Mô tả

API thiết kế dựa trên cách **client sử dụng**, không phải cấu trúc database hay logic nội bộ. API phải dễ hiểu, dễ dùng, ít gây nhầm lẫn.

### DO ✅

- Tên resource rõ ràng, gần với domain business
- Thiết kế theo use-case thực tế của client

```http
GET /users/123/orders
```

### DON'T ❌

- Lộ cấu trúc database hay tên bảng
- Dùng tên kỹ thuật khó hiểu

```http
GET /getUserOrderByUserId?id=123
```

---

## 2. Sử dụng chuẩn RESTful

### Mô tả

API tuân theo nguyên tắc REST: resource-based, HTTP methods đúng ngữ nghĩa, stateless.

### DO ✅

- Dùng HTTP method đúng mục đích: `GET` đọc, `POST` tạo, `PUT` cập nhật toàn bộ, `PATCH` cập nhật một phần, `DELETE` xoá
- URL đại diện cho resource, không phải action

```http
GET    /users        # Lấy danh sách
POST   /users        # Tạo mới
PUT    /users/123    # Cập nhật toàn bộ
DELETE /users/123    # Xoá
```

### DON'T ❌

- Dùng sai method (VD: `GET` để xoá dữ liệu)
- Đặt action vào URL

```http
POST /deleteUser?id=123
GET  /createUser?name=John
```

---

## 3. Command Pattern cho non-RESTful actions

### Mô tả

Khi hành động không thể biểu diễn bằng CRUD (GET/POST/PUT/DELETE), dùng **Command Pattern**: giữ prefix RESTful, thêm verb suffix mô tả action.

- Dùng `POST` cho action có side-effect, `GET` cho action chỉ đọc dữ liệu
- Prefix tuân thủ RESTful: `/api/v1/{resource}/{id}`
- Suffix là verb mô tả hành động: `cancel`, `approve`, `resend`, `do-something`

### DO ✅

- Đặt action verb ở cuối URL, sau resource path
- Dùng kebab-case cho action nhiều từ

```http
POST /api/v1/orders/123/cancel
POST /api/v1/orders/123/mark-as-paid
GET  /api/v1/reports/123/export
POST /api/v1/users/456/resend-verification
```

### DON'T ❌

- Đặt action ở đầu URL hoặc dùng query params
- Dùng method không phù hợp (VD: `GET` cho action có side-effect)

```http
POST /cancelOrder?id=123
GET  /api/v1/orders/123/cancel   # cancel có side-effect, không nên GET
POST /api/v1/cancel-order/123   # action không nên là resource name
```

---

## 4. Naming Convention rõ ràng, nhất quán

### Mô tả

Tên API dùng **danh từ số nhiều**, snake_case hoặc kebab-case, nhất quán toàn project.

### DO ✅

- Dùng noun (danh từ), số nhiều
- Giữ convention nhất quán

```http
GET /order-items
GET /products
```

### DON'T ❌

- Dùng verb trong URL
- Mix nhiều naming convention

```http
GET /getProduct
GET /order_items   # trong khi chỗ khác dùng kebab-case
```

---

## 5. Sử dụng HTTP Status Code đúng chuẩn

### Mô tả

Trả về status code phản ánh đúng kết quả xử lý. Client dựa vào status code để xử lý logic.

### DO ✅

- `200` thành công, `201` tạo mới thành công
- `400` bad request, `401` chưa xác thực, `403` không có quyền, `404` không tìm thấy
- `500` lỗi server

```http
POST /users → 201 Created
GET  /users/999 → 404 Not Found
```

### DON'T ❌

- Luôn trả `200` kèm error trong body
- Dùng status code không đúng ngữ nghĩa

```json
// Status 200 nhưng thực chất là lỗi
{
  "success": false,
  "error": "User not found"
}
```

---

## 6. Error Handling rõ ràng

### Mô tả

Response lỗi cần format nhất quán, có error code và message để client xử lý được.

### DO ✅

- Một format error duy nhất cho toàn project
- Có `code` (machine-readable) và `message` (human-readable)

```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User does not exist"
  }
}
```

### DON'T ❌

- Mỗi endpoint trả error format khác nhau
- Trả message mơ hồ hoặc thiếu error code

```json
{ "msg": "fail" }
{ "error": "Something went wrong" }
```

---

## 7. Pagination cho danh sách lớn

### Mô tả

Trả dữ liệu theo trang, tránh response quá lớn ảnh hưởng performance.

### DO ✅

- Hỗ trợ `page` + `limit` (hoặc `offset` + `limit`)
- Trả kèm thông tin pagination trong response

```http
GET /products?page=1&limit=10
```

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100
  }
}
```

### DON'T ❌

- Trả toàn bộ records không giới hạn
- Không cho client biết tổng số records

```http
GET /products
```

```json
{ "data": [...hàng nghìn items...] }
```

---

## 8. Filtering, Sorting, Searching

### Mô tả

Cho phép client query linh hoạt qua query params thay vì tạo nhiều endpoint riêng biệt.

### DO ✅

- Dùng query params cho filter, sort, search
- Đặt tên param rõ nghĩa

```http
GET /products?category=electronics&sort=price:asc&q=iphone
```

### DON'T ❌

- Tạo endpoint riêng cho mỗi filter
- Dùng request body cho GET request

```http
GET /products/electronics
GET /products/cheap
POST /products/search   # dùng POST thay vì GET query params
```

---

## 9. Idempotency

### Mô tả

Đảm bảo request gửi nhiều lần cho cùng kết quả, tránh duplicate data. Đặc biệt quan trọng với `POST`.

### DO ✅

- Dùng `Idempotency-Key` header cho POST request
- `PUT`, `DELETE` nên idempotent theo chuẩn

```http
POST /payments
Idempotency-Key: abc-123-xyz
```

### DON'T ❌

- POST không có cơ chế chống duplicate
- Tạo nhiều record giống nhau khi client retry

```http
POST /payments   # gửi 3 lần → tạo 3 payments
```

---

## 10. Versioning API

### Mô tả

Đánh version API để maintain backward compatibility khi thay đổi breaking changes.

### DO ✅

- Đặt version trong URL path
- Giữ version cũ hoạt động khi release version mới

```http
GET /v1/users
GET /v2/users
```

### DON'T ❌

- Thay đổi API mà không tăng version
- Dùng quá nhiều version không cần thiết

```http
GET /users   # thay đổi response format mà không báo trước
```

---

## 11. Consistent Response Format

### Mô tả

Toàn bộ API trả về cùng một cấu trúc response, giúp client parse dễ dàng.

### DO ✅

- Cấu trúc cố định: `data`, `meta`, `error`

```json
{
  "data": {},
  "meta": {},
  "error": null
}
```

### DON'T ❌

- Mỗi endpoint trả format khác nhau
- Trả raw value không wrap

```json
{}
[]
"ok"
```

---

## 12. Security Best Practices

### Mô tả

Bảo vệ API khỏi các lỗ hổng phổ biến: authentication, authorization, data exposure.

### DO ✅

- Dùng HTTPS cho mọi request
- Validate & sanitize input
- Dùng JWT/OAuth2, không expose sensitive data trong URL

```http
Authorization: Bearer <token>
```

### DON'T ❌

- Truyền token qua query params
- Trả về sensitive fields không cần thiết (password, secret)

```http
GET /users/123?token=abc123

# Response chứa password hash
{ "id": 123, "password": "$2b$..." }
```

---

## 13. Document rõ ràng

### Mô tả

API phải có tài liệu đầy đủ, cập nhật, để developer hiểu và tích hợp nhanh.

### DO ✅

- Dùng OpenAPI/Swagger để generate docs tự động
- Mỗi endpoint có mô tả, request/response example

```yaml
# Swagger annotation
/users:
  get:
    summary: Get all users
    responses:
      200:
        description: List of users
```

### DON'T ❌

- Không có document, chỉ truyền miệng
- Document outdated, không khớp API thực tế

```
# README.md viết 6 tháng trước, endpoint đã đổi
GET /user → thực tế đã đổi thành GET /v2/users
```