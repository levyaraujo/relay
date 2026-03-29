# Accessibility Checklist (WCAG 2.2 AA)

Relay-specific accessibility checklist. Every frontend PR should pass these checks.

---

## Perceivable

### Text & Content
- [ ] All text meets 4.5:1 contrast ratio against its background
- [ ] Large text (≥ 18px bold or ≥ 24px regular) meets 3:1 contrast
- [ ] UI components and graphical objects meet 3:1 contrast against adjacent colors
- [ ] Text can be resized to 200% without loss of content (no overflow: hidden on text)
- [ ] Content doesn't rely solely on color to convey information
  - Status badges: color + text label
  - Amount direction: color + plus/minus sign or debit/credit label
  - Form errors: color + icon + text message

### Images & Media
- [ ] All `<img>` have descriptive `alt` attributes
- [ ] Decorative images use `alt=""` and `aria-hidden="true"`
- [ ] Charts/graphs have text alternatives (data table or summary)
- [ ] Icons paired with text labels (or `aria-label` for icon-only buttons)

---

## Operable

### Keyboard
- [ ] All interactive elements focusable via Tab (in logical order)
- [ ] All actions triggerable via Enter or Space
- [ ] Visible focus indicator on every focusable element (never `outline: none` alone)
- [ ] Focus ring: 2px solid `--color-primary`, 2px offset
- [ ] No keyboard traps (except open modals — which trap intentionally)
- [ ] Skip-to-content link as first focusable element
- [ ] Dropdown menus navigable with arrow keys
- [ ] Modals: Escape to close, focus trapped inside, focus restored on close
- [ ] Custom components with roles (tabs, accordion) follow WAI-ARIA patterns

### Touch & Pointer
- [ ] All touch targets ≥ 44×44px (WCAG 2.5.8)
- [ ] Adequate spacing between touch targets (≥ 8px gap)
- [ ] No functionality requires multipoint or path-based gestures (pinch, swipe)
  as the only option — always provide a button alternative
- [ ] Drag-and-drop has a non-drag alternative (buttons to reorder)

### Timing
- [ ] No time limits on form completion
- [ ] Toast notifications persist long enough to read (≥ 4 seconds)
- [ ] Error toasts don't auto-dismiss
- [ ] `prefers-reduced-motion` respected — disable animations, transitions ≤ 0ms

---

## Understandable

### Language & Labels
- [ ] `<html lang="pt-BR">` set
- [ ] Form inputs have visible `<label>` elements (via `htmlFor` or wrapping)
- [ ] Labels are descriptive ("Valor da transacao" not just "Valor")
- [ ] Placeholder text supplements but never replaces labels
- [ ] Error messages explain what's wrong and how to fix it
  - Bad: "Campo invalido"
  - Good: "CPF deve ter 11 digitos"
- [ ] Helper text for non-obvious fields (e.g., "Informe o CNPJ sem pontos")

### Forms & Errors
- [ ] Required fields: mark optional with "(opcional)" rather than required with "*"
- [ ] Validation errors appear inline next to the field, not in a banner
- [ ] `aria-invalid="true"` on fields with errors
- [ ] `aria-describedby` links error message element to the input
- [ ] Error messages visible without scrolling (auto-scroll to first error)
- [ ] Form submission doesn't clear valid fields on error
- [ ] Success confirmation after form submission (toast or redirect)

### Navigation
- [ ] Consistent navigation across pages (same position, same order)
- [ ] Current page indicated in nav (aria-current="page")
- [ ] Page `<title>` is descriptive and unique per route
- [ ] Breadcrumbs on nested pages (using `<nav aria-label="Breadcrumb">`)

---

## Robust

### Semantic HTML
- [ ] Interactive elements use correct tags:
  - Actions → `<button>` (never `<div onClick>`)
  - Navigation → `<a href>` (never `<button>` for navigation)
  - Lists → `<ul>/<ol>` with `<li>`
  - Tables → `<table>` with `<thead>`, `<th scope>`, `<tbody>`
  - Forms → `<form>` with `<fieldset>` and `<legend>` for groups
- [ ] Landmark regions: `<header>`, `<nav>`, `<main>`, `<footer>`
- [ ] Only one `<main>` per page
- [ ] Headings follow hierarchy (h1 → h2 → h3, no skipping levels)
- [ ] One `<h1>` per page

### ARIA Usage
- [ ] ARIA used only when native HTML semantics are insufficient
- [ ] All `aria-*` attributes have valid values
- [ ] Dynamic content regions use `aria-live`:
  - Status updates, toast: `aria-live="polite"`
  - Urgent errors: `aria-live="assertive"` or `role="alert"`
- [ ] Loading states: `aria-busy="true"` on the loading container
- [ ] Custom widgets follow WAI-ARIA Authoring Practices patterns

### Relay-Specific
- [ ] Financial amounts announced correctly by screen readers
  (use `aria-label` if the visual format is ambiguous)
- [ ] Status transitions announced (`aria-live` on status badges if they change)
- [ ] Table responsive mode (card list on mobile) preserves data relationships
  (use `aria-label` or structured headings in cards)
- [ ] Currency inputs: screen reader announces "Reais" not "R dollar sign"
- [ ] Date inputs: use native `<input type="date">` which has built-in a11y

---

## Testing Tools

| Tool                     | What it catches                      |
|--------------------------|--------------------------------------|
| axe-core / Lighthouse    | Automated WCAG violations            |
| Keyboard-only testing    | Focus order, traps, missing handlers |
| Chrome DevTools a11y tab | ARIA tree, contrast, names           |
| VoiceOver / NVDA         | Screen reader experience             |
| zoom to 200%             | Text reflow, overflow issues         |
| prefers-reduced-motion   | Animation respect                    |
