# Component Patterns Reference

Detailed specs for relay's shared UI components.

---

## Button

```tsx
type ButtonProps = {
  variant: "primary" | "secondary" | "danger" | "ghost"
  size: "sm" | "md" | "lg"
  loading?: boolean
  disabled?: boolean
  fullWidth?: boolean
  children: React.ReactNode
  onClick?: () => void
  type?: "button" | "submit" | "reset"
}
```

Rules:
- Always has visible text (icon-only buttons need `aria-label`)
- `loading` state disables the button and shows a spinner inside
- `danger` variant requires confirmation before destructive action
- Full-width on mobile for primary actions

Sizes:
- `sm`: 32px height, 12px padding, `sm` font — table row actions
- `md`: 40px height, 16px padding, `base` font — default
- `lg`: 48px height, 20px padding, `lg` font — primary CTA, mobile

---

## Input

```tsx
type InputProps = {
  label: string
  name: string
  type?: "text" | "number" | "email" | "tel" | "date" | "password"
  inputMode?: "text" | "decimal" | "numeric" | "tel" | "email"
  placeholder?: string
  helperText?: string
  error?: string
  required?: boolean
  disabled?: boolean
  value: string
  onChange: (value: string) => void
  mask?: "cpf" | "cnpj" | "phone" | "currency"
}
```

Rules:
- Label is always visible (never rely on placeholder alone)
- Error message replaces helper text when present
- Error state: red border + red error text + `aria-invalid="true"`
- `aria-describedby` links to helper/error text element
- Currency inputs: right-aligned, `inputMode="decimal"`, format on blur

Layout:
```
[Label]                     ← always above, `sm` font, `text-secondary`
[Input field]               ← `base` font, 40px height, 12px padding
[Helper text or error]      ← `xs` font, below input
```

---

## Table

Two rendering modes based on viewport:

### Desktop (≥ md): Standard table
```tsx
<table>
  <thead> (sticky)
  <tbody>
    <tr> per row, <td> per cell
  </tbody>
</table>
```

### Mobile (< md): Card list
```tsx
<ul>
  <li> per item
    <div> primary info (description, amount)
    <div> secondary info (date, status badge)
  </li>
</ul>
```

Props:
```tsx
type Column<T> = {
  key: keyof T
  label: string
  align?: "left" | "right"
  sortable?: boolean
  hideBelow?: "sm" | "md" | "lg"  // responsive column hiding
  render?: (value: T[keyof T], row: T) => React.ReactNode
}

type TableProps<T> = {
  columns: Column<T>[]
  data: T[]
  loading?: boolean
  emptyMessage?: string
  emptyCTA?: { label: string; onClick: () => void }
  onRowClick?: (row: T) => void
  sortBy?: keyof T
  sortDirection?: "asc" | "desc"
  onSort?: (key: keyof T) => void
}
```

---

## Badge (Status Pill)

```tsx
type BadgeProps = {
  variant: "neutral" | "success" | "warning" | "danger" | "info"
  children: string
}
```

Mapping for financial statuses:
- `draft` → neutral (gray)
- `posted` / `active` → info (blue)
- `paid` / `settled` → success (green)
- `overdue` → danger (red)
- `pending` → warning (orange)
- `void` / `cancelled` → neutral (gray, strikethrough)

Always render as a `<span>` with `role="status"` if dynamically updated.

---

## Modal

```tsx
type ModalProps = {
  open: boolean
  onClose: () => void
  title: string
  children: React.ReactNode
  actions?: React.ReactNode  // footer buttons
}
```

Rules:
- Trap focus inside when open
- Close on Escape key
- Close on backdrop click
- Return focus to trigger element on close
- Use `<dialog>` element (native) when possible
- Backdrop: semi-transparent overlay
- Max width: 480px, centered vertically
- On mobile: full-screen bottom sheet (slides up from bottom)

Use sparingly — only for confirmations and quick forms. Never for complex workflows.

---

## Card

```tsx
type CardProps = {
  children: React.ReactNode
  padding?: "sm" | "md" | "lg"
  onClick?: () => void        // makes it a clickable card
  highlighted?: boolean       // subtle background color
}
```

- Default padding: `md` (16px)
- Border: 1px solid `--color-border`, 8px border-radius
- If clickable: cursor pointer, hover shadow, focus ring
- Used for: dashboard widgets, mobile list items, summary blocks

---

## PageHeader

```tsx
type PageHeaderProps = {
  title: string
  subtitle?: string
  action?: React.ReactNode    // primary action button
  backTo?: string             // shows back arrow, navigates to this path
}
```

Layout:
```
[← Back]                           (if backTo provided)
[Title]                [Action Button]
[Subtitle]
```

- Title: `xl` font
- Subtitle: `sm` font, `text-secondary`
- On mobile: action button goes full-width below title

---

## Toast / Notification

```tsx
type ToastProps = {
  variant: "success" | "error" | "warning" | "info"
  message: string
  duration?: number  // ms, default 4000, 0 = persistent
}
```

- Position: top-right on desktop, top-center on mobile
- Auto-dismiss after duration (except errors: persist until dismissed)
- `aria-live="polite"` container
- `role="alert"` for error toasts
- Max 3 visible at once, stack vertically
- Dismissible via close button or swipe (mobile)

---

## Skeleton Loader

Match the shape of the content being loaded:
- Text lines: rounded rectangles, varying widths (100%, 80%, 60%)
- Amounts: right-aligned rectangle
- Avatar/icon: circle
- Table: repeat row skeleton 5 times

Use CSS animation (`@keyframes shimmer`) — no JS needed.

```css
.skeleton {
  background: linear-gradient(90deg,
    var(--color-bg-muted) 25%,
    var(--color-bg-subtle) 50%,
    var(--color-bg-muted) 75%
  );
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
```
