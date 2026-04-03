# Vielish UI Redesign — Design Spec

**Date:** 2026-04-03
**Scope:** Frontend visual overhaul — design system, landing page, auth pages, dashboard stats, flashcard TTS
**Approach:** Phased — design system first, then apply per-page

---

## Goals

- Establish a consistent warm/friendly visual identity across all pages
- Fix theme inconsistency (currently: dark landing, light dashboard)
- Add meaningful stats to dashboard so users know what to do each day
- Add TTS pronunciation to flashcard

## Out of Scope

- New pages or routes
- Backend changes (except `/api/stats` endpoint dependency noted below)
- OAuth / social login
- Flashcard flip animation, swipe gestures, keyboard shortcuts
- Feature sections or marketing content on landing page

---

## Phase 1: Design System

Define color tokens in `tailwind.config.ts` and use them consistently across all pages.

### Color Palette

| Token | Value | Usage |
|-------|-------|-------|
| `warm-bg` | `#fffbeb` | Page backgrounds |
| `warm-surface` | `#ffffff` | Cards, inputs, modals |
| `warm-border` | `#fde68a` | Borders, dividers |
| `warm-accent` | `#f59e0b` | Primary CTA, active states, highlights |
| `warm-accent-hover` | `#d97706` | Button hover state |
| `warm-text` | `#1c1917` | Headings, primary text |
| `warm-muted` | `#78350f` | Secondary text, subtitles |
| `warm-subtle` | `#a16207` | Captions, labels, placeholders |

**Typography:** Keep Inter (Next.js default). No additional fonts.
**Spacing:** Keep Tailwind defaults.

### Tailwind Config

Add the tokens under `theme.extend.colors.warm` in `tailwind.config.ts`.

---

## Phase 2: Landing Page (`/`)

**Goal:** Convert visitors with a clear value proposition. Keep minimal — no feature sections, no illustrations.

### Layout

```
[Navbar: logo left | Đăng nhập outline + Đăng ký solid right]

[Centered hero]
  Heading:    "Học từ vựng tiếng Anh — không bao giờ quên"
  Subheading: "SRS thông minh · Giao diện Việt · Miễn phí"
  CTAs:       [Đăng ký miễn phí →] (amber solid)  [Đăng nhập] (outline)

[Social proof line]
  "⭐ Được dùng bởi hàng nghìn học viên Việt Nam"
  (placeholder copy — update when real user numbers are available)
```

### Changes from Current

- Background: `black` → `warm-bg` (`#fffbeb`)
- Logo color: white → `#92400e`
- Add simple navbar (logo + 2 auth buttons)
- Replace single-line tagline with stronger heading + sub-heading
- Replace 2 plain buttons with amber solid CTA + outline secondary
- Add social proof line below CTAs

---

## Phase 3: Auth Pages (`/login`, `/register`)

**Goal:** Consistent with warm theme. No functional changes.

### Changes from Current

- Background: `black` → `warm-bg`
- Input border: white/gray → `warm-border` (`#fde68a`), focus ring `warm-accent`
- Submit button: blue → amber (`warm-accent`)
- Link color: blue → `#b45309`
- Add small "Vielish" logo above form, links back to `/`

---

## Phase 4: Dashboard (`/dashboard`)

**Goal:** Give users immediate context — what's their streak, how many words learned, what needs reviewing today.

### Layout

```
[Greeting]
  "Xin chào, [display_name]! 👋"
  (display_name from useAuth() context — the user's display name set at registration)

[Stats row — 3 cards]
  🔥 [streak] ngày streak
  📚 [total_learned] từ đã học
  🔔 [due_today] từ cần ôn hôm nay

[Action cards — 2 cards]
  🃏 Chủ đề từ vựng         |  🔁 Ôn tập hôm nay
  "Học từ mới theo chủ đề"  |  "[N] từ đang chờ bạn" (if due > 0)
                             |  badge highlight if due > 0
```

### Navbar

Add active state to nav links: amber underline on current route (use `usePathname()`).

### API Dependency

Stats require a new endpoint: `GET /api/stats`

Response shape:
```json
{
  "streak": 7,
  "total_learned": 42,
  "due_today": 5
}
```

**Graceful degradation:** If the endpoint is unavailable or returns an error, display `–` in stat cards. Do not block dashboard render.

---

## Phase 5: Flashcard TTS (`/components/flashcard.tsx`)

**Goal:** Let users hear pronunciation without leaving the flashcard.

### Implementation

Add a 🔊 icon button in the top-right corner of the flashcard (visible on both front and back faces).

```ts
const speak = (text: string) => {
  const utterance = new SpeechSynthesisUtterance(text)
  utterance.lang = 'en-US'
  window.speechSynthesis.speak(utterance)
}
```

- Clicking the button calls `speak(word.word)`
- No loading state needed (Web Speech API is synchronous)
- **Fallback:** If `!window.speechSynthesis`, hide the button silently — no error message

### No Backend Required

Uses the browser's built-in Web Speech API. No TTS endpoint, no API key.

---

## Implementation Order

1. `tailwind.config.ts` — add warm color tokens
2. `/` — landing page warm theme + new copy + navbar
3. `/login`, `/register` — apply warm theme
4. `/dashboard` — greeting + stats row + action card improvements + nav active state
5. `GET /api/stats` backend endpoint (can be stub returning zeros initially)
6. `/components/flashcard.tsx` — TTS button

Each phase is independently deployable.

---

## Implementation Plans

- [x] **Plan 1: UI Redesign** — Warm theme, dashboard stats, flashcard TTS across all pages
  - `docs/agent-docs/plans/2026-04-03-ui-redesign.md`
