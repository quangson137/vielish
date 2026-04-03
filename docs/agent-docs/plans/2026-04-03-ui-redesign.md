# UI Redesign Implementation Plan

> Steps use checkbox (`- [x]`) syntax for tracking progress.

**Goal:** Apply a consistent warm/friendly theme across all pages, add dashboard stats, and TTS pronunciation to flashcards.

**Architecture:** Frontend-first approach — define CSS color tokens in Tailwind v4's `@theme inline`, then update each page's JSX/classes. One new backend endpoint (`GET /api/stats`) provides dashboard stats. Auth context extended to expose `displayName` from localStorage.

**Tech Stack:** Next.js (App Router), Tailwind CSS v4, Go (Gin), GORM, PostgreSQL, Web Speech API

---

### Task 1: Design System — Warm Color Tokens

**Files:**
- Modify: `web/app/globals.css`

- [x] **Step 1: Update CSS variables and theme tokens**

Replace the entire `web/app/globals.css` with warm color tokens:

```css
@import "tailwindcss";

:root {
  --background: #fffbeb;
  --foreground: #1c1917;
  --warm-bg: #fffbeb;
  --warm-surface: #ffffff;
  --warm-border: #fde68a;
  --warm-accent: #f59e0b;
  --warm-accent-hover: #d97706;
  --warm-text: #1c1917;
  --warm-muted: #78350f;
  --warm-subtle: #a16207;
}

@theme inline {
  --color-background: var(--background);
  --color-foreground: var(--foreground);
  --color-warm-bg: var(--warm-bg);
  --color-warm-surface: var(--warm-surface);
  --color-warm-border: var(--warm-border);
  --color-warm-accent: var(--warm-accent);
  --color-warm-accent-hover: var(--warm-accent-hover);
  --color-warm-text: var(--warm-text);
  --color-warm-muted: var(--warm-muted);
  --color-warm-subtle: var(--warm-subtle);
  --font-sans: var(--font-geist-sans);
  --font-mono: var(--font-geist-mono);
}

body {
  background: var(--background);
  color: var(--foreground);
  font-family: Arial, Helvetica, sans-serif;
}
```

Note: Removed the `prefers-color-scheme: dark` media query — Vielish is warm-only, no dark mode toggle.

- [x] **Step 2: Verify the dev server loads without errors**

Run: `cd web && npm run dev`
Expected: App loads at localhost:3000. Background is now cream (#fffbeb) instead of white/dark.

---

### Task 2: Landing Page Redesign

**Files:**
- Modify: `web/app/page.tsx`

- [x] **Step 1: Replace landing page with warm minimal layout**

Replace `web/app/page.tsx` entirely:

```tsx
import Link from "next/link";

export default function Home() {
  return (
    <div className="min-h-screen bg-warm-bg">
      <nav className="flex items-center justify-between px-6 py-4">
        <span className="text-xl font-bold text-warm-muted">Vielish</span>
        <div className="flex gap-3">
          <Link
            href="/login"
            className="px-4 py-2 text-sm border border-warm-accent text-warm-accent rounded-lg hover:bg-warm-accent hover:text-white transition-colors"
          >
            Đăng nhập
          </Link>
          <Link
            href="/register"
            className="px-4 py-2 text-sm bg-warm-accent text-white rounded-lg hover:bg-warm-accent-hover transition-colors"
          >
            Đăng ký
          </Link>
        </div>
      </nav>

      <main className="flex flex-col items-center justify-center px-8 py-32">
        <h1 className="text-4xl font-bold text-warm-text mb-4 text-center">
          Học từ vựng tiếng Anh — không bao giờ quên
        </h1>
        <p className="text-lg text-warm-muted mb-8 text-center">
          SRS thông minh · Giao diện Việt · Miễn phí
        </p>
        <div className="flex gap-4">
          <Link
            href="/register"
            className="px-6 py-3 bg-warm-accent text-white rounded-lg hover:bg-warm-accent-hover font-semibold transition-colors"
          >
            Đăng ký miễn phí →
          </Link>
          <Link
            href="/login"
            className="px-6 py-3 border border-warm-accent text-warm-accent rounded-lg hover:bg-warm-accent hover:text-white transition-colors"
          >
            Đăng nhập
          </Link>
        </div>
        <p className="mt-12 text-warm-subtle text-sm border-t border-dashed border-warm-border pt-4">
          ⭐ Được dùng bởi hàng nghìn học viên Việt Nam
        </p>
      </main>
    </div>
  );
}
```

- [x] **Step 2: Verify landing page renders correctly**

Run: Open `http://localhost:3000` in browser.
Expected: Cream background, navbar with logo + 2 buttons, centered hero with tagline, 2 CTAs, social proof line. All in amber/warm color scheme.

---

### Task 3: Auth Pages — Warm Theme

**Files:**
- Modify: `web/components/auth-form.tsx`
- Modify: `web/app/login/page.tsx`
- Modify: `web/app/register/page.tsx`

- [x] **Step 1: Update AuthForm component with warm colors**

In `web/components/auth-form.tsx`, replace the following class strings:

1. Error div: change `bg-red-50 text-red-700` → keep as-is (red for errors is fine)
2. All input elements: change `focus:ring-blue-500` → `focus:ring-warm-accent`

   Replace all 3 occurrences of:
   ```
   className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
   ```
   with:
   ```
   className="w-full px-3 py-2 border border-warm-border rounded-lg focus:outline-none focus:ring-2 focus:ring-warm-accent bg-warm-surface"
   ```

3. Submit button: change `bg-blue-600 ... hover:bg-blue-700` → amber:

   Replace:
   ```
   className="w-full py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
   ```
   with:
   ```
   className="w-full py-3 bg-warm-accent text-white rounded-lg hover:bg-warm-accent-hover disabled:opacity-50"
   ```

- [x] **Step 2: Update login page**

Replace `web/app/login/page.tsx`:

```tsx
"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import AuthForm from "@/components/auth-form";
import { useAuth } from "@/lib/auth-context";

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuth();

  const handleLogin = async (data: {
    email: string;
    password: string;
  }) => {
    await login(data.email, data.password);
    router.push("/dashboard");
  };

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8 bg-warm-bg">
      <Link href="/" className="text-2xl font-bold text-warm-muted mb-8 hover:text-warm-accent transition-colors">
        Vielish
      </Link>
      <h1 className="text-3xl font-bold mb-8 text-warm-text">Đăng nhập</h1>
      <AuthForm mode="login" onSubmit={handleLogin} />
      <p className="mt-4 text-sm text-warm-muted">
        Chưa có tài khoản?{" "}
        <Link href="/register" className="text-warm-accent hover:underline">
          Đăng ký
        </Link>
      </p>
    </main>
  );
}
```

- [x] **Step 3: Update register page**

Read `web/app/register/page.tsx` and apply the same pattern as login:
- Add `bg-warm-bg` to `<main>`
- Add Vielish logo link to `/`
- Change link color to `text-warm-accent`
- Change text colors to warm tokens

Replace `web/app/register/page.tsx`:

```tsx
"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import AuthForm from "@/components/auth-form";
import { useAuth } from "@/lib/auth-context";

export default function RegisterPage() {
  const router = useRouter();
  const { register } = useAuth();

  const handleRegister = async (data: {
    email: string;
    password: string;
    displayName?: string;
  }) => {
    await register(data.email, data.password, data.displayName || "");
    router.push("/dashboard");
  };

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8 bg-warm-bg">
      <Link href="/" className="text-2xl font-bold text-warm-muted mb-8 hover:text-warm-accent transition-colors">
        Vielish
      </Link>
      <h1 className="text-3xl font-bold mb-8 text-warm-text">Tạo tài khoản</h1>
      <AuthForm mode="register" onSubmit={handleRegister} />
      <p className="mt-4 text-sm text-warm-muted">
        Đã có tài khoản?{" "}
        <Link href="/login" className="text-warm-accent hover:underline">
          Đăng nhập
        </Link>
      </p>
    </main>
  );
}
```

- [x] **Step 4: Verify auth pages**

Run: Open `http://localhost:3000/login` and `http://localhost:3000/register` in browser.
Expected: Cream background, amber buttons, warm-colored inputs. Vielish logo links back to `/`.

---

### Task 4: Auth Context — Expose displayName

**Files:**
- Modify: `web/lib/auth-context.tsx`

- [x] **Step 1: Add displayName to auth context**

The auth context needs to expose `displayName` for the dashboard greeting. Since the backend login response doesn't include it, store it in localStorage during register. For login users, it will be empty.

Replace `web/lib/auth-context.tsx`:

```tsx
"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from "react";
import { api } from "./api";

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  displayName: string;
  login: (email: string, password: string) => Promise<void>;
  register: (
    email: string,
    password: string,
    displayName: string
  ) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [displayName, setDisplayName] = useState("");

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    const name = localStorage.getItem("display_name") || "";
    setIsAuthenticated(!!token);
    setDisplayName(name);
    setIsLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    await api.login(email, password);
    setIsAuthenticated(true);
  };

  const register = async (
    email: string,
    password: string,
    displayName: string
  ) => {
    await api.register(email, password, displayName);
    localStorage.setItem("display_name", displayName);
    setDisplayName(displayName);
    setIsAuthenticated(true);
  };

  const logout = () => {
    api.clearTokens();
    localStorage.removeItem("display_name");
    setIsAuthenticated(false);
    setDisplayName("");
  };

  return (
    <AuthContext.Provider
      value={{ isAuthenticated, isLoading, displayName, login, register, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
```

- [x] **Step 2: Verify auth context compiles**

Run: `cd web && npx tsc --noEmit`
Expected: No type errors. The new `displayName` field is available in `useAuth()`.

---

### Task 5: Backend — Stats Endpoint

**Files:**
- Modify: `server/internal/domain/vocab/repository.go`
- Modify: `server/internal/driven/vocab/repository.go`
- Modify: `server/internal/appcore/vocab/dto.go`
- Modify: `server/internal/appcore/vocab/usecase.go`
- Modify: `server/internal/driving/httpui/handler/vocab_handler.go`
- Modify: `server/internal/driving/httpui/presenter/vocab_presenter.go`
- Modify: `server/internal/driving/httpui/server.go`
- Create: `server/internal/appcore/vocab/streak_test.go` (streak calculation test)

- [x] **Step 1: Add repository interface methods**

In `server/internal/domain/vocab/repository.go`, add 3 new methods to the `Repository` interface:

```go
type Repository interface {
	// Topics
	ListTopics(ctx context.Context, level string) ([]Topic, error)
	GetTopicByID(ctx context.Context, id string) (*Topic, error)

	// Words
	ListWordsByTopic(ctx context.Context, topicID string) ([]Word, error)
	GetWordByID(ctx context.Context, id string) (*Word, error)
	GetRandomWords(ctx context.Context, topicID string, excludeID string, limit int) ([]Word, error)

	// Progress
	GetProgress(ctx context.Context, userID, wordID string) (*UserWordProgress, error)
	UpsertProgress(ctx context.Context, progress *UserWordProgress) error
	GetDueWords(ctx context.Context, userID string, now time.Time, limit int) ([]Word, error)

	// Stats
	CountLearnedWords(ctx context.Context, userID string) (int, error)
	CountDueWords(ctx context.Context, userID string, now time.Time) (int, error)
	GetReviewDates(ctx context.Context, userID string) ([]time.Time, error)
}
```

- [x] **Step 2: Add StatsOutput DTO**

In `server/internal/appcore/vocab/dto.go`, add at the end:

```go
type StatsOutput struct {
	Streak       int `json:"streak"`
	TotalLearned int `json:"total_learned"`
	DueToday     int `json:"due_today"`
}
```

- [x] **Step 3: Add GetStats use case method**

In `server/internal/appcore/vocab/usecase.go`, add:

```go
func (uc *UseCase) GetStats(ctx context.Context, userID string) (*StatsOutput, error) {
	totalLearned, err := uc.repo.CountLearnedWords(ctx, userID)
	if err != nil {
		return nil, err
	}

	dueToday, err := uc.repo.CountDueWords(ctx, userID, time.Now())
	if err != nil {
		return nil, err
	}

	dates, err := uc.repo.GetReviewDates(ctx, userID)
	if err != nil {
		return nil, err
	}

	streak := calculateStreak(dates)

	return &StatsOutput{
		Streak:       streak,
		TotalLearned: totalLearned,
		DueToday:     dueToday,
	}, nil
}

func calculateStreak(dates []time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	streak := 0
	today := time.Now().Truncate(24 * time.Hour)

	// dates are expected in descending order (most recent first)
	expected := today
	for _, d := range dates {
		day := d.Truncate(24 * time.Hour)
		if day.Equal(expected) {
			streak++
			expected = expected.AddDate(0, 0, -1)
		} else if day.Before(expected) {
			break
		}
	}

	return streak
}
```

- [x] **Step 4: Write test for calculateStreak**

Create `server/internal/appcore/vocab/streak_test.go`:

```go
package appcore

import (
	"testing"
	"time"
)

func TestCalculateStreak(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)

	tests := []struct {
		name   string
		dates  []time.Time
		expect int
	}{
		{
			name:   "no reviews",
			dates:  nil,
			expect: 0,
		},
		{
			name:   "reviewed today only",
			dates:  []time.Time{today.Add(10 * time.Hour)},
			expect: 1,
		},
		{
			name: "3 consecutive days",
			dates: []time.Time{
				today.Add(8 * time.Hour),
				today.AddDate(0, 0, -1).Add(14 * time.Hour),
				today.AddDate(0, 0, -2).Add(9 * time.Hour),
			},
			expect: 3,
		},
		{
			name: "gap breaks streak",
			dates: []time.Time{
				today.Add(8 * time.Hour),
				today.AddDate(0, 0, -2).Add(14 * time.Hour),
			},
			expect: 1,
		},
		{
			name: "no review today",
			dates: []time.Time{
				today.AddDate(0, 0, -1).Add(14 * time.Hour),
				today.AddDate(0, 0, -2).Add(9 * time.Hour),
			},
			expect: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateStreak(tt.dates)
			if got != tt.expect {
				t.Errorf("calculateStreak() = %d, want %d", got, tt.expect)
			}
		})
	}
}
```

- [x] **Step 5: Run streak test**

Run: `cd server && go test ./internal/appcore/vocab/ -run TestCalculateStreak -v`
Expected: All 5 sub-tests PASS.

- [x] **Step 6: Implement repository methods**

In `server/internal/driven/vocab/repository.go`, add these 3 methods:

```go
func (r *Repository) CountLearnedWords(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&UserWordProgressModel{}).
		Where("user_id = ? AND review_count > 0", userID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("counting learned words: %w", err)
	}
	return int(count), nil
}

func (r *Repository) CountDueWords(ctx context.Context, userID string, now time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&UserWordProgressModel{}).
		Where("user_id = ? AND next_review_at <= ?", userID, now).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("counting due words: %w", err)
	}
	return int(count), nil
}

func (r *Repository) GetReviewDates(ctx context.Context, userID string) ([]time.Time, error) {
	var dates []time.Time
	err := r.db.WithContext(ctx).
		Model(&UserWordProgressModel{}).
		Where("user_id = ? AND last_reviewed_at IS NOT NULL", userID).
		Select("DISTINCT DATE(last_reviewed_at) as review_date").
		Order("review_date DESC").
		Limit(90).
		Pluck("review_date", &dates).Error
	if err != nil {
		return nil, fmt.Errorf("getting review dates: %w", err)
	}
	return dates, nil
}
```

- [x] **Step 7: Add handler interface method and handler**

In `server/internal/driving/httpui/handler/vocab_handler.go`, add `GetStats` to the interface:

```go
type VocabUseCaseInterface interface {
	ListTopics(ctx context.Context, level string) ([]appcore.TopicOutput, error)
	GetTopicWords(ctx context.Context, topicID string) ([]appcore.WordOutput, error)
	GetWord(ctx context.Context, id string) (*appcore.WordOutput, error)
	GetDueReviews(ctx context.Context, userID string) ([]appcore.WordOutput, error)
	SubmitReview(ctx context.Context, userID, wordID string, input appcore.ReviewInput) error
	GetQuiz(ctx context.Context, topicID string) ([]appcore.QuizQuestion, error)
	SubmitQuiz(ctx context.Context, userID, topicID string, input appcore.QuizAnswerInput) (*appcore.QuizResult, error)
	GetStats(ctx context.Context, userID string) (*appcore.StatsOutput, error)
}
```

Add the handler method at the end of the file:

```go
func (h *VocabHandler) GetStats(c *gin.Context) {
	userID, ok := ctxbase.GetUserID(c.Request.Context())
	if !ok {
		httpbase.Error(c, http.StatusUnauthorized, "user not found in context")
		return
	}
	stats, err := h.useCase.GetStats(c.Request.Context(), userID)
	if err != nil {
		httpbase.Error(c, http.StatusInternalServerError, "failed to get stats")
		return
	}
	h.presenter.Stats(c, http.StatusOK, stats)
}
```

- [x] **Step 8: Add presenter method**

In `server/internal/driving/httpui/presenter/vocab_presenter.go`, add:

```go
func (p *VocabPresenter) Stats(c *gin.Context, status int, stats *appcore.StatsOutput) {
	httpbase.Success(c, status, stats)
}
```

- [x] **Step 9: Register the route**

In `server/internal/driving/httpui/server.go`, inside the `protected` group (after the quiz routes), add:

```go
protected.GET("/stats", vocabHandler.GetStats)
```

The full protected block becomes:

```go
protected := r.Group("/api").Use(middleware.Auth(svc))
{
	protected.GET("/review/due", vocabHandler.GetDueReviews)
	protected.POST("/review/:wordId", vocabHandler.SubmitReview)
	protected.GET("/quiz/:topicId", vocabHandler.GetQuiz)
	protected.POST("/quiz/:topicId", vocabHandler.SubmitQuiz)
	protected.GET("/stats", vocabHandler.GetStats)
}
```

- [x] **Step 10: Verify backend compiles**

Run: `cd server && go build ./cmd/api/`
Expected: Build succeeds with no errors.

---

### Task 6: Frontend — Stats API Client

**Files:**
- Modify: `web/lib/vocab-api.ts`

- [x] **Step 1: Add fetchStats function**

Add at the end of `web/lib/vocab-api.ts`:

```typescript
export interface UserStats {
  streak: number;
  total_learned: number;
  due_today: number;
}

export async function fetchStats(): Promise<UserStats> {
  const res = await api.request("/api/stats");
  if (!res.ok) throw new Error("Failed to fetch stats");
  return res.json();
}
```

---

### Task 7: Dashboard — Stats + Action Cards

**Files:**
- Modify: `web/app/dashboard/page.tsx`
- Modify: `web/app/dashboard/layout.tsx`

- [x] **Step 1: Rewrite dashboard page with greeting, stats, and action cards**

Replace `web/app/dashboard/page.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useAuth } from "@/lib/auth-context";
import { fetchStats, UserStats } from "@/lib/vocab-api";

export default function DashboardPage() {
  const { displayName } = useAuth();
  const [stats, setStats] = useState<UserStats | null>(null);

  useEffect(() => {
    fetchStats()
      .then(setStats)
      .catch(console.error);
  }, []);

  const greeting = displayName
    ? `Xin chào, ${displayName}! 👋`
    : "Xin chào! 👋";

  return (
    <div>
      <h2 className="text-2xl font-bold mb-6 text-warm-text">{greeting}</h2>

      <div className="grid grid-cols-3 gap-4 mb-8">
        <div className="bg-warm-surface border border-warm-border rounded-lg p-4 text-center">
          <p className="text-3xl font-bold text-warm-accent">
            {stats?.streak ?? "–"}
          </p>
          <p className="text-sm text-warm-muted">ngày streak 🔥</p>
        </div>
        <div className="bg-warm-surface border border-warm-border rounded-lg p-4 text-center">
          <p className="text-3xl font-bold text-green-600">
            {stats?.total_learned ?? "–"}
          </p>
          <p className="text-sm text-warm-muted">từ đã học 📚</p>
        </div>
        <div className="bg-warm-surface border border-warm-border rounded-lg p-4 text-center">
          <p className="text-3xl font-bold text-blue-600">
            {stats?.due_today ?? "–"}
          </p>
          <p className="text-sm text-warm-muted">cần ôn hôm nay 🔔</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Link
          href="/dashboard/topics"
          className="block p-6 bg-warm-accent rounded-lg hover:bg-warm-accent-hover transition-colors text-white"
        >
          <p className="text-2xl mb-1">🃏</p>
          <h3 className="text-lg font-semibold mb-1">Chủ đề từ vựng</h3>
          <p className="text-sm opacity-80">
            Học từ mới theo chủ đề với flashcard và SRS.
          </p>
        </Link>
        <Link
          href="/dashboard/review"
          className="block p-6 bg-warm-surface border border-warm-border rounded-lg hover:shadow-md transition-shadow"
        >
          <p className="text-2xl mb-1">🔁</p>
          <h3 className="text-lg font-semibold text-warm-text mb-1">
            Ôn tập hôm nay
          </h3>
          <p className="text-sm text-warm-muted">
            {stats && stats.due_today > 0
              ? `${stats.due_today} từ đang chờ bạn`
              : "Ôn lại các từ đã học theo lịch SRS."}
          </p>
        </Link>
      </div>
    </div>
  );
}
```

- [x] **Step 2: Update dashboard layout with warm theme and nav active state**

Replace `web/app/dashboard/layout.tsx`:

```tsx
"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/lib/auth-context";

const navLinks = [
  { href: "/dashboard", label: "Trang chủ" },
  { href: "/dashboard/topics", label: "Chủ đề" },
  { href: "/dashboard/review", label: "Ôn tập" },
];

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isLoading, logout } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push("/login");
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-warm-bg">
        <p className="text-warm-muted">Đang tải...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-warm-bg">
      <nav className="border-b border-warm-border px-6 py-4 flex justify-between items-center bg-warm-surface">
        <div className="flex items-center gap-6">
          <h1 className="text-xl font-bold text-warm-muted">Vielish</h1>
          {navLinks.map((link) => {
            const isActive =
              link.href === "/dashboard"
                ? pathname === "/dashboard"
                : pathname.startsWith(link.href);
            return (
              <Link
                key={link.href}
                href={link.href}
                className={`text-sm transition-colors ${
                  isActive
                    ? "text-warm-accent font-semibold border-b-2 border-warm-accent pb-1"
                    : "text-warm-muted hover:text-warm-text"
                }`}
              >
                {link.label}
              </Link>
            );
          })}
        </div>
        <button
          onClick={() => {
            logout();
            router.push("/login");
          }}
          className="text-sm text-warm-muted hover:text-warm-text"
        >
          Đăng xuất
        </button>
      </nav>
      <main className="p-6">{children}</main>
    </div>
  );
}
```

- [x] **Step 3: Verify dashboard**

Run: Register a new account and navigate to `/dashboard`.
Expected: Greeting with display name, 3 stat cards (showing "–" if backend stats endpoint is not running), 2 action cards with icons and warm colors. Nav has active state highlighting.

---

### Task 8: Flashcard — TTS Button

**Files:**
- Modify: `web/components/flashcard.tsx`

- [x] **Step 1: Add TTS support to flashcard**

Replace `web/components/flashcard.tsx`:

```tsx
"use client";

import { useState, useEffect } from "react";
import { Word } from "@/lib/vocab-api";

interface FlashcardProps {
  word: Word;
  onRate: (quality: 1 | 3 | 5) => void;
}

export default function Flashcard({ word, onRate }: FlashcardProps) {
  const [flipped, setFlipped] = useState(false);
  const [hasTTS, setHasTTS] = useState(false);

  useEffect(() => {
    setHasTTS(typeof window !== "undefined" && "speechSynthesis" in window);
  }, []);

  useEffect(() => {
    setFlipped(false);
  }, [word]);

  const speak = () => {
    const utterance = new SpeechSynthesisUtterance(word.word);
    utterance.lang = "en-US";
    window.speechSynthesis.speak(utterance);
  };

  return (
    <div className="max-w-md mx-auto">
      <div
        onClick={() => setFlipped(!flipped)}
        className="relative border border-warm-border bg-warm-surface rounded-lg p-8 min-h-[250px] flex flex-col items-center justify-center cursor-pointer hover:shadow-md transition-shadow"
      >
        {hasTTS && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              speak();
            }}
            className="absolute top-4 right-4 text-warm-subtle hover:text-warm-accent transition-colors text-xl"
            aria-label="Phát âm"
          >
            🔊
          </button>
        )}

        {!flipped ? (
          <>
            <p className="text-3xl font-bold mb-2 text-warm-text">{word.word}</p>
            {word.ipa_phonetic && (
              <p className="text-warm-subtle">{word.ipa_phonetic}</p>
            )}
            {word.part_of_speech && (
              <p className="text-warm-subtle text-sm mt-1">
                {word.part_of_speech}
              </p>
            )}
            <p className="text-warm-subtle text-sm mt-4">Nhấn để lật thẻ</p>
          </>
        ) : (
          <>
            <p className="text-2xl font-semibold text-warm-accent mb-3">
              {word.vi_meaning}
            </p>
            {word.en_definition && (
              <p className="text-warm-muted text-sm mb-2">
                {word.en_definition}
              </p>
            )}
            {word.example_sentence && (
              <div className="mt-3 text-sm text-warm-subtle">
                <p className="italic">{word.example_sentence}</p>
                {word.example_vi_translation && (
                  <p className="text-warm-subtle mt-1">
                    {word.example_vi_translation}
                  </p>
                )}
              </div>
            )}
          </>
        )}
      </div>

      {flipped && (
        <div className="flex justify-center gap-3 mt-4">
          <button
            onClick={() => onRate(1)}
            className="px-6 py-2 bg-red-500 text-white rounded hover:bg-red-600"
          >
            Khó
          </button>
          <button
            onClick={() => onRate(3)}
            className="px-6 py-2 bg-yellow-500 text-white rounded hover:bg-yellow-600"
          >
            Ổn
          </button>
          <button
            onClick={() => onRate(5)}
            className="px-6 py-2 bg-green-500 text-white rounded hover:bg-green-600"
          >
            Dễ
          </button>
        </div>
      )}
    </div>
  );
}
```

Key changes from original:
- Added `hasTTS` state (checks `window.speechSynthesis` on mount)
- Added `speak()` function using Web Speech API
- Added 🔊 button in top-right corner, hidden if no TTS support
- `e.stopPropagation()` prevents flip when clicking TTS
- Added `useEffect` to reset `flipped` when `word` changes (bug fix)
- Applied warm color tokens to all text

- [x] **Step 2: Verify TTS works**

Run: Navigate to a learn page and click the 🔊 button.
Expected: Browser speaks the English word aloud. Button hidden in browsers without Speech API.

---

### Task 9: Final Verification

- [x] **Step 1: Full visual check**

Open each page in browser and verify warm theme is consistent:

1. `http://localhost:3000` — cream bg, amber CTAs, social proof
2. `http://localhost:3000/login` — cream bg, amber button, Vielish logo
3. `http://localhost:3000/register` — same as login + display name field
4. `http://localhost:3000/dashboard` — greeting, 3 stats, 2 action cards, nav active state
5. `http://localhost:3000/dashboard/topics` — warm bg inherited from layout
6. Flashcard page — warm card, TTS button visible

- [x] **Step 2: Run backend tests**

Run: `cd server && go test ./...`
Expected: All tests pass including the new streak tests.

- [x] **Step 3: Run frontend type check**

Run: `cd web && npx tsc --noEmit`
Expected: No type errors.
