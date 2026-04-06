# Logging Standard

## 1. Không log sensitive info

Tuyệt đối không log các thông tin nhạy cảm: password, token, secret key, số thẻ, OTP.

**DON'T**

```ts
logger.info('User login', { email, password }); // ❌ Lộ password
logger.info(`Token: ${accessToken}`);            // ❌ Lộ token
```

**DO**

```ts
logger.info('User login', { email });            // ✅ Chỉ log thông tin cần thiết
logger.info('Token issued', { userId });         // ✅ Log context, không log giá trị nhạy cảm
```

---

## 2. Chỉ log ở biên (middleware), không log trong core function

Log request/response tại middleware (entry point). Các hàm logic bên trong **không tự log** — chỉ throw error hoặc return kết quả.

**DON'T**

```ts
// core function tự log → log bị trùng lặp, khó kiểm soát
async function createOrder(data: CreateOrderDto) {
  logger.info('Creating order', data);           // ❌ Log trong core
  const order = await this.orderRepo.save(data);
  logger.info('Order created', order);           // ❌ Log trong core
  return order;
}
```

**DO**

```ts
// --- Core function: thuần logic, không log ---
async function createOrder(data: CreateOrderDto) {
  const order = await this.orderRepo.save(data); // ✅ Chỉ làm business logic
  return order;                                   // ✅ Throw error nếu lỗi, middleware sẽ bắt
}
```

> **Tại sao?** Khi log tập trung ở biên, ta kiểm soát được format, tránh log trùng, và dễ dàng thay đổi log level hoặc transport mà không sửa business logic.

---

## 3. Log phải có trace_id (OpenTelemetry)

Mỗi dòng log phải gắn `trace_id` từ OpenTelemetry để có thể trace toàn bộ request xuyên suốt các service.

**DON'T**

```ts
logger.info('Order created', { orderId });       // ❌ Không có trace_id → không trace được
```

**DO**

```ts
import { trace } from '@opentelemetry/api';

// Lấy trace_id từ span hiện tại
const span = trace.getActiveSpan();
const traceId = span?.spanContext().traceId;

logger.info({ traceId, orderId, message: 'Order created' }); // ✅
```

> **Tip:** Cấu hình logger tự động inject `trace_id` vào mọi log entry thay vì truyền tay (xem ví dụ tổng hợp bên dưới).

---

## 4. Log phải có format JSON

Dùng structured JSON log — **không dùng plain text**. JSON giúp các hệ thống log aggregation (ELK, Loki, Datadog) parse và query dễ dàng.

**DON'T**

```ts
logger.info(`[${new Date()}] Order created: ${orderId}`); // ❌ Plain text, không parse được
```

**DO**

```json
{
  "timestamp": "2026-04-03T09:00:00.000Z",
  "level": "info",
  "trace_id": "abc123def456",
  "message": "Order created",
  "orderId": "order_789"
}
```

---

## 5. Error log phải có stacktrace

Khi log error, **bắt buộc** phải kèm stacktrace gốc để dev có thể trỏ thẳng đến dòng code gây lỗi.

**DON'T**

```ts
logger.error('Something went wrong');              // ❌ Không biết lỗi ở đâu
logger.error({ message: error.message });          // ❌ Mất stacktrace
```

**DO**

```ts
logger.error({
  message: error.message,
  stack: error.stack,                              // ✅ Có stacktrace đầy đủ
  orderId,
});
```

---

## Ví dụ tổng hợp — Middleware logging đầy đủ

```ts
import { trace } from '@opentelemetry/api';
import pino from 'pino';

// --- Logger: format JSON, tự động gắn trace_id ---
const logger = pino({
  formatters: {
    log(obj) {
      const span = trace.getActiveSpan();
      return {
        ...obj,
        trace_id: span?.spanContext().traceId ?? 'N/A',
      };
    },
  },
});

// --- Middleware: log request + response ---
app.use((req, res, next) => {
  const start = Date.now();

  res.on('finish', () => {
    logger.info({
      message: 'HTTP Request',
      method: req.method,
      path: req.originalUrl,
      status: res.statusCode,
      duration: Date.now() - start,
    });
  });

  next();
});

// --- Error handler: log error + stacktrace ---
app.use((err, req, res, next) => {
  logger.error({
    message: err.message,
    stack: err.stack,
    method: req.method,
    path: req.originalUrl,
  });

  res.status(err.status ?? 500).json({ error: 'Internal Server Error' });
});
```

**Output mẫu — success:**

```json
{
  "level": 30,
  "timestamp": "2026-04-03T09:00:00.000Z",
  "trace_id": "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6",
  "message": "HTTP Request",
  "method": "POST",
  "path": "/api/orders",
  "status": 201,
  "duration": 45
}
```

**Output mẫu — error:**

```json
{
  "level": 50,
  "timestamp": "2026-04-03T09:00:01.000Z",
  "trace_id": "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6",
  "message": "Cannot read property 'id' of undefined",
  "stack": "TypeError: Cannot read property 'id' of undefined\n    at createOrder (src/orders/order.service.ts:42:18)\n    at OrderController.create (src/orders/order.controller.ts:15:30)",
  "method": "POST",
  "path": "/api/orders"
}
```
