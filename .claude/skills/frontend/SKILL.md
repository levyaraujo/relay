---
name: frontend
description: >
  Frontend UI/UX guideline for the relay ERP project. Use this skill whenever
  building, designing, or reviewing frontend components — pages, forms, tables,
  navigation, accessibility, responsive layout, color, typography, or any
  user-facing interface work. Triggers on: new page, new component, build UI,
  add form, design layout, accessibility, a11y, responsive, style, CSS, frontend,
  React component, UX review, or any task involving frontend/ files.
---

# Frontend UI/UX Guideline

Design and implementation reference for the **relay** ERP frontend.
Stack: React 19 · TypeScript · Vite · CSS (no UI framework — keep it lean).

Target audience: **micro entrepreneurs** in Brazil — people who are not tech-savvy,
often working from a phone or a cheap laptop, with limited time and patience for
complex interfaces.

---

## Design Philosophy

### 1. Radical Simplicity

Micro entrepreneurs are not accountants. They want to:
- Record what they earned and what they spent
- Know if they can pay their bills this month
- Send an invoice / boleto
- See if customers owe them money

Every screen must answer **one question** clearly. If a page tries to do two things,
split it. Complexity is the enemy — relay wins by being the ERP that doesn't feel
like an ERP.

### 2. Mobile-First, Always

Over 60% of micro entrepreneurs in Brazil access business tools from their phone.
Design for a 360px viewport first, then scale up.

```
Breakpoints:
  sm:  ≥ 480px   (large phone)
  md:  ≥ 768px   (tablet)
  lg:  ≥ 1024px  (laptop)
  xl:  ≥ 1280px  (desktop)
```

- Touch targets: minimum 44×44px (WCAG 2.5.8)
- No hover-only interactions — everything must work on touch
- Avoid horizontal scrolling at all costs
- Forms should stack vertically on mobile — never side-by-side fields below `md`

### 3. Progressive Disclosure

Don't show everything at once. Use:
- **Summary → detail** patterns (list view → detail view)
- **Expandable sections** for advanced options
- **Smart defaults** so users rarely need to change settings
- **Inline help** (short tooltip or helper text) instead of documentation pages

---

## Visual Design

### Color Palette

Keep it minimal. Financial software should feel trustworthy, not playful.

```css
/* Semantic tokens — define in CSS custom properties */
--color-bg:            #FFFFFF;
--color-bg-subtle:     #F8F9FA;
--color-bg-muted:      #F1F3F5;
--color-text:          #212529;
--color-text-secondary:#868E96;
--color-border:        #DEE2E6;

--color-primary:       #228BE6;   /* actions, links, focus rings   */
--color-primary-hover: #1C7ED6;
--color-success:       #40C057;   /* positive amounts, confirmations */
--color-danger:        #FA5252;   /* negative amounts, destructive actions */
--color-warning:       #FD7E14;   /* overdue, attention needed */

/* Dark mode (optional, later) */
--color-bg-dark:       #1A1B1E;
--color-text-dark:     #C1C2C5;
```

- Use color **sparingly** — mainly for status indicators and CTAs
- Financial amounts: green for income/credit, red for expense/debit — universally understood
- Never rely on color alone to convey meaning (a11y)

### Typography

One typeface. System font stack for zero load time:

```css
--font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
             "Helvetica Neue", Arial, sans-serif;
--font-mono: "SF Mono", "Fira Code", "Fira Mono", Menlo, monospace;
```

Type scale (modular, base 16px):

| Token    | Size   | Weight | Use                         |
|----------|--------|--------|-----------------------------|
| `xs`     | 12px   | 400    | Captions, helper text       |
| `sm`     | 14px   | 400    | Secondary text, table cells |
| `base`   | 16px   | 400    | Body text, inputs           |
| `lg`     | 18px   | 500    | Section headers             |
| `xl`     | 24px   | 600    | Page titles                 |
| `2xl`    | 30px   | 700    | Dashboard big numbers       |

- Line height: 1.5 for body, 1.25 for headings
- Financial numbers: always use `font-variant-numeric: tabular-nums` for alignment
- Currency: always display with `R$` prefix and 2 decimal places (`R$ 1.250,00`)
- Use Brazilian number format: period for thousands, comma for decimals

### Spacing

Use a 4px base unit. Common spacing values:

```
4px  (0.25rem) — tight inner padding
8px  (0.5rem)  — default gap between related elements
12px (0.75rem) — input padding
16px (1rem)    — card padding, section gap
24px (1.5rem)  — between sections
32px (2rem)    — page-level vertical rhythm
48px (3rem)    — major section separation
```

### Icons

Use a single icon set (Lucide or Phosphor — both are small and tree-shakeable).
Icons are supplementary — always pair with a text label, except in compact
table actions where a tooltip is sufficient.

---

## Component Patterns

### Forms

Forms are the core interaction in an ERP. Get them right.

```
Rules:
1. One column on mobile, two columns on md+ only when fields are related
   (e.g., city + state)
2. Labels above inputs (not floating, not placeholder-only)
3. Required fields: mark optional fields with "(optional)" — most fields
   are required, so marking required adds noise
4. Validation: inline, on blur, with clear error message below the field
5. Submit button: always visible (not hidden behind scroll), full-width on mobile
6. Destructive actions: require confirmation (modal or inline "are you sure?")
7. Auto-save drafts when possible — micro entrepreneurs get interrupted constantly
```

Field sizing:
- Amounts: right-aligned, monospace, `inputmode="decimal"`
- Dates: native `<input type="date">` — no custom date pickers on mobile
- CPF/CNPJ: use `inputmode="numeric"` and auto-format with mask
- Phone: `inputmode="tel"`

### Tables / Lists

Financial data is often tabular. But tables are hard on mobile.

```
Rules:
1. Below md breakpoint: switch to card/list layout (one item per card)
2. Above md: use a proper <table> with sticky header
3. Always show: description, amount, date, status — hide secondary
   columns on smaller screens
4. Sortable columns: date and amount at minimum
5. Amounts: right-aligned, monospace, colored by type (green/red)
6. Status: use a small colored badge (pill) — not just text
7. Empty state: friendly message + CTA ("No transactions yet. Record your first one.")
8. Loading: skeleton placeholders, never a full-page spinner
```

### Navigation

Micro entrepreneurs don't need a 20-item sidebar.

```
Mobile:  Bottom tab bar (max 4-5 items) — Dashboard, Transactions, Invoices, More
Tablet+: Collapsible sidebar with icon + label
```

Key nav items (in priority order):
1. **Dashboard** — overview, cash balance, recent activity
2. **Transactions** — record income/expenses (the core action)
3. **Invoices** — send/track invoices
4. **Reports** — P&L, cash flow
5. **Settings** — company, bank accounts, users

### Dashboard

The dashboard answers: "How is my business doing right now?"

Must show:
- Current cash balance (big number, prominent)
- Income vs expenses this month (simple bar or comparison)
- Overdue receivables (how much customers owe, with count)
- Upcoming payables (what's due soon)
- Recent transactions (last 5-10, quick glance)

Don't show: charts that require financial literacy, KPIs that need explanation.

---

## Accessibility (A11Y)

relay must be usable by everyone. Follow WCAG 2.2 AA as baseline.

### Non-Negotiable Rules

1. **Semantic HTML** — use `<button>` for actions, `<a>` for navigation,
   `<table>` for tabular data, `<form>` for forms. No `<div onClick>`.
2. **Keyboard navigation** — every interactive element reachable via Tab,
   activated via Enter/Space. Visible focus ring (`outline`) on all focusable
   elements. Never `outline: none` without a replacement.
3. **Color contrast** — minimum 4.5:1 for normal text, 3:1 for large text
   (WCAG AA). Use a contrast checker for all color combinations.
4. **Labels** — every `<input>` must have an associated `<label>` (via `htmlFor`
   or wrapping). Screen readers need this.
5. **Alt text** — every `<img>` has descriptive `alt`. Decorative images use
   `alt=""` and `aria-hidden="true"`.
6. **ARIA** — use ARIA only when semantic HTML isn't enough. Prefer native
   elements. Common needs:
   - `aria-label` for icon-only buttons
   - `aria-live="polite"` for status messages and toast notifications
   - `aria-describedby` for error messages linked to inputs
   - `role="alert"` for important error notifications
7. **Motion** — respect `prefers-reduced-motion`. No auto-playing animations.
   Keep transitions under 200ms.
8. **Language** — set `lang="pt-BR"` on `<html>`. Screen readers use this for
   pronunciation.
9. **Touch targets** — minimum 44×44px. Add padding if the visual element is
   smaller.
10. **Error handling** — errors must be announced to screen readers. Use
    `aria-invalid="true"` on fields with errors + `aria-describedby` pointing
    to the error message element.

### Testing Accessibility

- Use browser DevTools Accessibility panel
- Test with keyboard only (no mouse)
- Run axe-core or Lighthouse accessibility audit
- Test with VoiceOver (Mac) or NVDA (Windows) at least once per major feature

---

## Component Architecture

### File Structure

```
frontend/src/
├── components/          # Shared, reusable UI components
│   ├── Button.tsx
│   ├── Input.tsx
│   ├── Table.tsx
│   ├── Card.tsx
│   ├── Badge.tsx
│   ├── Modal.tsx
│   └── Layout/
│       ├── Sidebar.tsx
│       ├── BottomNav.tsx
│       └── PageHeader.tsx
├── pages/               # Route-level page components
│   ├── Dashboard.tsx
│   ├── Transactions.tsx
│   ├── TransactionDetail.tsx
│   └── Settings.tsx
├── hooks/               # Custom React hooks
│   ├── useApi.ts        # Data fetching
│   └── useMediaQuery.ts # Responsive breakpoints
├── utils/               # Pure utility functions
│   ├── format.ts        # Currency, date, CPF/CNPJ formatting
│   └── validation.ts    # Client-side validation helpers
├── styles/              # Global CSS
│   └── tokens.css       # CSS custom properties (colors, spacing, fonts)
├── App.tsx
└── main.tsx
```

### Rules

1. **Components are functions** — no class components.
2. **Props over context** for component configuration. Context for truly global
   state (auth, theme, locale).
3. **No prop drilling** deeper than 2 levels — introduce context or composition.
4. **Colocation** — styles, tests, and types live next to the component that
   uses them.
5. **No barrel files** (`index.ts` re-exports) — import directly from the
   source file. Barrel files create circular dependencies and slow builds.
6. **CSS approach** — CSS Modules or plain CSS with BEM-like naming. No CSS-in-JS
   runtime (keep the bundle small). Utility classes only for spacing/layout.
7. **State management** — React state + context. No Redux or external state
   library unless complexity demands it. Start simple.

---

## Data Fetching

Keep it simple. `fetch` + a thin custom hook.

```typescript
// hooks/useApi.ts — minimal data fetching hook
function useApi<T>(url: string) {
  const [data, setData] = useState<T | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch(url)
      .then(res => {
        if (!res.ok) throw new Error(res.statusText)
        return res.json()
      })
      .then(setData)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [url])

  return { data, error, loading }
}
```

- Always handle loading, error, and empty states
- Show skeleton loaders, not spinners
- Optimistic updates for simple mutations (toggle, delete)
- Debounce search inputs (300ms)

---

## Formatting Standards (Brazilian Locale)

These are critical — financial data must be formatted correctly.

```typescript
// utils/format.ts

// Currency: R$ 1.234,56
function formatCurrency(value: number): string {
  return value.toLocaleString("pt-BR", {
    style: "currency",
    currency: "BRL",
  })
}

// Date: 23/03/2026
function formatDate(date: Date | string): string {
  return new Date(date).toLocaleDateString("pt-BR")
}

// CPF: 123.456.789-09
function formatCPF(cpf: string): string {
  return cpf.replace(/(\d{3})(\d{3})(\d{3})(\d{2})/, "$1.$2.$3-$4")
}

// CNPJ: 12.345.678/0001-90
function formatCNPJ(cnpj: string): string {
  return cnpj.replace(
    /(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})/,
    "$1.$2.$3/$4-$5"
  )
}
```

---

## Performance Budget

Micro entrepreneurs are often on slow connections (3G/4G). Keep the bundle small.

```
Targets:
- First Contentful Paint:  < 1.5s on 4G
- Total JS bundle:         < 150KB gzipped (initial load)
- Largest Contentful Paint: < 2.5s
- No layout shift (CLS):   < 0.1
```

Rules:
- Lazy-load routes with `React.lazy` + `Suspense`
- No heavy charting libraries on initial load — lazy-load if needed
- Optimize images: use WebP, proper sizing, `loading="lazy"`
- Prefer CSS over JS for animations
- Tree-shake everything — check bundle with `vite-bundle-visualizer`

---

## Error & Empty States

Every data-driven view needs three states beyond the happy path:

| State   | What to show                                            |
|---------|---------------------------------------------------------|
| Loading | Skeleton placeholder matching the layout shape          |
| Empty   | Friendly message + primary action CTA                   |
| Error   | What went wrong + what the user can do (retry, contact) |

Empty state messages should be encouraging, not technical:
- "Nenhuma transacao ainda. Registre sua primeira receita ou despesa."
- "Sem faturas pendentes. Tudo em dia!"

Error messages should be human:
- "Nao conseguimos carregar seus dados. Tente novamente."
- Never show raw error codes or stack traces

---

## Checklist Before Shipping a Page

- [ ] Works on 360px viewport (test in Chrome DevTools)
- [ ] All interactive elements reachable via keyboard (Tab through the page)
- [ ] Color contrast passes WCAG AA (4.5:1 for text)
- [ ] All inputs have visible labels
- [ ] Loading, empty, and error states implemented
- [ ] Financial amounts formatted in BRL (`R$ 1.234,56`)
- [ ] Dates formatted in pt-BR (`DD/MM/YYYY`)
- [ ] Touch targets ≥ 44×44px
- [ ] No horizontal scroll on mobile
- [ ] `lang="pt-BR"` set on html element
- [ ] Page has a clear, descriptive `<title>`
- [ ] Runs Lighthouse accessibility audit ≥ 90

---

## Detailed References

- `references/component-patterns.md` — Detailed component API specs, props, and usage examples
- `references/a11y-checklist.md` — Expanded WCAG 2.2 AA checklist with relay-specific examples