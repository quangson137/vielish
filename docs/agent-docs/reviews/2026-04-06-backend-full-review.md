# Code Review: Full Backend Server Review

**Date:** 2026-04-06
**Reviewed by:** Claude (reviewing-code skill)
**Scope:** All Go source files in `server/` (~40 files across domain, appcore, driven, driving, and pkg layers)
**Change type:** Full codebase review

---

## Summary

The Vielish backend is a well-architected Go application following hexagonal/clean architecture with clear layer separation (domain, appcore, driven, driving). The codebase demonstrates good practices: proper DI via Uber Fx, parameterized SQL queries, JWT auth with refresh token rotation, and solid test coverage. However, there are several security issues (missing rate limiting, CSRF on logout, hardcoded dev secret), a few correctness bugs (streak timezone handling, quiz endpoint auth), and areas where error handling and observability could be strengthened.

---

## Findings

### 🔴 Critical - Must Fix

- [ ] `server/internal/driving/httpui/server.go:60` **GetQuiz endpoint is public but should be protected** - `GET /api/quiz/:topicId` is inside the `protected` group, but `POST /quiz/:topicId` (SubmitQuiz) accesses `userID` from context. However, `GetQuiz` at line 60 does not use userID so it technically works, but the quiz generation is only useful for authenticated users. More critically: verify that the route grouping is intentional. Currently quiz GET/POST are both in protected, which is correct. *(After re-reading: this is actually fine. Downgrading.)*

- [ ] `server/cmd/api/config.yaml:12` **JWT secret hardcoded in committed config** - `jwt.secret: dev-secret-change-in-production` is committed to the repo. While env var override is supported via Viper, this default is dangerously weak (28 bytes) and could be accidentally used in production if `JWT_SECRET` env var is not set. **Suggested fix:** Remove the default secret from config.yaml, require it via env var, and fail fast on startup if `jwt.secret` is empty in production mode.

- [ ] `server/internal/driving/httpui/server.go:27-63` **No rate limiting on auth endpoints** - `/api/auth/login`, `/api/auth/register`, and `/api/auth/refresh` have no rate limiting, making them vulnerable to brute-force and credential stuffing attacks. **Suggested fix:** Add a rate-limiting middleware (e.g., `gin-contrib/limiter` or a Redis-based limiter) on the auth group.

- [ ] `server/internal/appcore/vocab/usecase.go:62-66` **`GetDueReviews` uses `time.Now()` directly, making it untestable and timezone-dependent** - The `time.Now()` call inside the use case means the behavior cannot be deterministically tested and the "due" calculation depends on server timezone. **Suggested fix:** Accept a `now time.Time` parameter or inject a clock interface.

### 🟡 Important - Should Fix

- [ ] `server/internal/appcore/vocab/usecase.go:161-180` **`calculateStreak` uses `time.Now()` — timezone-sensitive and untestable from GetStats** - The streak calculation truncates to midnight using server's local timezone, but review dates come from PostgreSQL which may use UTC. A user reviewing at 11pm local time might not get credit if the server runs in UTC. **Suggested fix:** Use a consistent timezone (UTC) throughout, or accept `now` as a parameter.

- [ ] `server/internal/driving/httpui/handler/vocab_handler.go:117` **Internal error message leaked to client** - `err.Error()` is returned directly in the response for `SubmitReview` failures. If an unexpected DB error occurs, this could leak internal details (table names, query info). **Suggested fix:** Map known errors to user-friendly messages, return generic 500 for unknown errors.

- [ ] `server/internal/driving/httpui/server.go:47` **Logout endpoint does not require authentication** - `/api/auth/logout` is outside the protected middleware group, meaning anyone with a valid refresh token string can call it. While the impact is limited (can only delete the token), it's inconsistent — logout should require the access token to prevent random token deletion. **Suggested fix:** Move logout to the protected group or verify the access token before accepting logout.

- [ ] `server/internal/appcore/vocab/usecase.go:90-119` **Quiz generation fetches ALL words in a topic** - `GetQuiz` loads every word in a topic and generates a question for each. For large topics this could be very slow and return huge payloads. **Suggested fix:** Add pagination or limit the number of quiz questions (e.g., max 10-20 per quiz).

- [ ] `server/internal/driven/vocab/repository.go:76-91` **`ORDER BY RANDOM()` is slow on large tables** - `GetRandomWords` uses `ORDER BY RANDOM()` which performs a full table scan. For MVP with small data this is fine, but will become a bottleneck as data grows. **Suggested fix:** For now, document this as a known limitation. Consider `TABLESAMPLE` or application-level random selection later.

- [ ] `server/internal/appcore/vocab/usecase.go:90-119` **N+1 query in GetQuiz** - For each word in a topic, `GetRandomWords` makes a separate DB query to fetch distractors. With 20 words in a topic, that's 21 queries. **Suggested fix:** Fetch all words once, then select distractors in-memory.

- [ ] `server/internal/domain/auth/service.go:20-24` **No validation of JWT secret length** - An empty or very short JWT secret would silently produce weak tokens. **Suggested fix:** Validate that `jwtSecret` is at least 32 bytes in `NewService`, return an error or panic if not.

- [ ] `server/internal/driven/database/redis.go:26` **Redis connection not verified on startup** - `log.Info("connected to redis")` is logged before actually pinging Redis. The client is created but connection is lazy. **Suggested fix:** Add `client.Ping(ctx)` in an `OnStart` hook to fail fast if Redis is unreachable.

- [ ] `server/internal/driven/database/gorm.go:13-19` **No connection pool configuration** - The GORM connection uses defaults for max open connections, idle connections, and connection lifetime. Under load this could exhaust database connections. **Suggested fix:** Expose pool settings in config (`max_open_conns`, `max_idle_conns`, `conn_max_lifetime`) and configure via `db.DB()`.

- [ ] `server/internal/driving/httpui/server.go:67-87` **No read/write timeouts on HTTP server** - The `http.Server` has no `ReadTimeout`, `WriteTimeout`, or `ReadHeaderTimeout`, making it vulnerable to slowloris-style attacks. **Suggested fix:** Set `ReadTimeout: 10s`, `WriteTimeout: 30s`, `ReadHeaderTimeout: 5s`.

### 🟢 Nice-to-Have - Consider

- [ ] `server/internal/driving/httpui/server.go:20-25` **No request logging middleware** - The logging standard doc specifies "log at boundaries" with trace IDs, but no request logging middleware is registered. `gin.New()` without `gin.Logger()` means no request logs. **Suggested fix:** Add a structured logging middleware that logs method, path, status, latency, and trace_id per the logging standard.

- [ ] `server/internal/driving/httpui/handler/auth_handler.go:56` **Validation error details exposed to client** - `"invalid input: " + err.Error()` returns Gin binding error details which may reveal internal field names and validation rules. **Suggested fix:** Return a generic "invalid input" message or map field errors to user-friendly messages.

- [ ] `server/pkg/httpbase/httpbase.go:5-11` **Response format inconsistency** - `Success` returns raw data, while `Error` wraps in `{"error": msg}`. The API design standard calls for a consistent format with `data` wrapper. Topics/Words/Word return raw arrays/objects, but errors wrap in `{"error": ...}`. **Suggested fix:** Wrap success responses in `{"data": ...}` for consistency with the documented standard.

- [ ] `server/internal/appcore/vocab/usecase.go:93-94` **`reviewRequest.Quality` binding tag `required` won't catch quality=0** - Gin's `binding:"required"` on an `int` field won't catch `0` as invalid because Go zero-value for int is 0. Since valid values are 1, 3, 5, this is handled by the `validQualities` map in the use case, so it works correctly. But the binding tag is misleading. **Suggested fix:** Use `binding:"required,oneof=1 3 5"` in the handler's `reviewRequest` struct.

- [ ] `server/internal/appcore/vocab/usecase.go:122-145` **SubmitQuiz does not update SRS progress** - Submitting a quiz returns scores but doesn't trigger SRS review updates for each answer. Users who use quiz mode won't have their progress tracked. **Suggested fix:** Consider calling `SubmitReview` for each quiz answer to update SRS data.

- [ ] `server/internal/domain/auth/entity.go:10` **Level field in User entity has no validation** - The `Level` field accepts any string. **Suggested fix:** Define level constants (`beginner`, `intermediate`, `advanced`) and validate on creation.

- [ ] `server/internal/driving/httpui/module.go` **No request ID middleware** - The `ctxbase` package defines `SetRequestID`/`GetRequestID` but no middleware generates or injects request IDs. **Suggested fix:** Add a middleware that generates a UUID request ID per request for tracing.

---

## Anti-Patterns Detected

| Pattern | Location | Severity |
|---|---|---|
| **N+1 Query** | `appcore/vocab/usecase.go:90-119` (GetQuiz) | 🟡 |
| **Time coupling** | `appcore/vocab/usecase.go:62,155` (time.Now() in business logic) | 🟡 |
| **Missing fail-fast** | `driven/database/redis.go:26` (no ping), `domain/auth/service.go:20` (no secret validation) | 🟡 |

---

## What's Done Well

- **Clean architecture discipline** — The hexagonal architecture is well-executed. Domain layer has zero infrastructure imports. Dependencies flow inward correctly. The `fx.Module` composition in `main.go` is clean and readable.

- **Proper auth token rotation** — The refresh token flow correctly deletes the old token before issuing a new one, preventing token reuse attacks. Refresh tokens stored in Redis with TTL is a solid approach.

- **Parameterized queries throughout** — All GORM queries use parameterized conditions (`Where("email = ?", email)`), preventing SQL injection. The `UpsertProgress` using GORM's `clause.OnConflict` is clean.

- **Good test coverage** — Domain services, use cases, and handlers all have unit tests with meaningful assertions. The mock/stub pattern for repository interfaces is well-implemented. The SM-2 algorithm tests verify EaseFactor math precisely.

- **Proper error domain boundaries** — Domain errors (`ErrUserNotFound`, `ErrInvalidCredentials`) are defined at the domain level and mapped to HTTP status codes at the handler level, keeping layers decoupled.

- **Interface-based testing** — Handler tests use `UseCaseInterface` / `VocabUseCaseInterface` for dependency injection, making them truly unit tests without infrastructure dependencies.

- **CORS properly configured** — CORS origins are configurable via config, not hardcoded. Credentials support is enabled for cookie/auth header scenarios.

---

## Checklist

- [x] Architecture fits existing patterns
- [ ] No security vulnerabilities (rate limiting, JWT secret validation, HTTP timeouts needed)
- [ ] Performance is acceptable (N+1 in quiz, ORDER BY RANDOM)
- [x] Tests cover new/changed logic
- [ ] Documentation is updated (logging middleware not implemented per standard)

---

## Verdict

**Request Changes**

The architecture and code quality are strong for an MVP. However, the missing rate limiting on auth endpoints, hardcoded JWT secret in config, and lack of HTTP server timeouts are security concerns that should be addressed before any production deployment. The N+1 query in quiz generation and timezone issues in streak calculation are important correctness/performance issues to fix soon.
