# v2 color system (Radix scale)

The v2 UI kit uses [Radix Colors](https://www.radix-ui.com/colors) 12-step scales as its color primitive. Each hue provides 12 numbered steps designed for specific use cases. Components consume these through Tailwind utility classes (`bg-sand-3`, `text-red-11`, `border-gold-7`, …).

Theme file: [`packages/ui/src/v2/theme.css`](../../packages/ui/src/v2/theme.css)

## Available scales

| Scale | Role |
|-------|------|
| **sand** | Neutral — primary UI chrome (backgrounds, borders, text) |
| **gold** | Warm accent neutral |
| **red** | Destructive / error |
| **green** | Success / positive |
| **amber** | Warning |
| **sky** | Informational |

## Step-to-usage mapping

Every scale follows the same 12-step structure:

| Step | Use case | Tailwind example |
|------|----------|------------------|
| 1 | App background | `bg-sand-1` |
| 2 | Subtle background | `bg-sand-2` |
| 3 | UI element background | `bg-sand-3` |
| 4 | Hovered UI element background | `hover:bg-sand-4` |
| 5 | Active / selected UI element background | `bg-sand-5` |
| 6 | Subtle borders and separators | `border-sand-6` |
| 7 | UI element border and focus rings | `border-sand-7` |
| 8 | Hovered UI element border | `border-sand-8` |
| 9 | Solid backgrounds | `bg-green-9` |
| 10 | Hovered solid backgrounds | `hover:bg-green-10` |
| 11 | Low-contrast text | `text-sand-11` |
| 12 | High-contrast text | `text-sand-12` |

### Quick mental model

Three bands: **low = light/background**, **middle = borders**, **high = text/solid**.

- **1–2** → backgrounds
- **3–5** → component backgrounds (normal → hover → active)
- **6–8** → borders (subtle → default → strong)
- **9–10** → solid backgrounds (normal → hover)
- **11–12** → text (low-contrast → high-contrast)

## Choosing a color step

1. **What am I styling?**
   - Background → steps 1–5 (or 9–10 for solid fills)
   - Border → steps 6–8
   - Text / icon → steps 11–12
2. **What state?**
   - Default → lower step in the range (3, 6, 9, 11)
   - Hover → next step up (4, 7, 10)
   - Active / pressed → one more (5, 8)
3. **Which hue?**
   - Neutral UI → `sand`
   - Semantic meaning → `red` (error), `green` (success), `amber` (warning), `sky` (info)
   - Warm accent → `gold`

## Neutral vs accent

Use **sand** for all general UI chrome: page backgrounds, card backgrounds, borders, primary text. Use hue scales only when conveying semantic meaning:

```tsx
// Neutral card
<div className="rounded-lg border border-sand-6 bg-sand-2 p-4">
  <p className="text-sand-12">Title</p>
  <p className="text-sand-11">Description</p>
</div>

// Error state
<div className="rounded-lg border border-red-6 bg-red-3 p-4">
  <p className="text-red-11">Something went wrong</p>
</div>

// Success badge
<span className="rounded bg-green-3 px-2 py-0.5 text-green-11">Approved</span>
```

## Contrast guarantees

Per the Radix spec, steps 11 and 12 are guaranteed to meet APCA contrast requirements on top of a step 1 or 2 background from the same scale. This means `text-sand-11` on `bg-sand-2` is always readable, and `text-sand-12` on `bg-sand-1` is always readable.

## Dark mode

**Never apply dark-mode color overrides in components.** The v2 theme imports `@radix-ui/colors` CSS files which handle light/dark switching automatically. Dark mode activates when a `.dark` class is present on `<html>`:

```ts
document.documentElement.classList.toggle("dark", isDark);
```

The same Tailwind classes (`bg-sand-1`, `text-red-11`, etc.) resolve to the correct dark values automatically because the `@theme inline` mappings reference the Radix variables (`var(--sand-1)`, etc.) which switch based on the `.dark` class. P3 wide-gamut colors are included for both light and dark modes on supported displays.

This is independent of v1's dark mode which uses `@variant dark` / `prefers-color-scheme`.

## Isolation from v1

v2 is a standalone theme isolated at the build level, not via a runtime DOM scope. An app opts into v2 by importing the v2 theme instead of the v1 `theme.css`:

```css
/* app index.css — v2 build */
@import "tailwindcss";
@import "@probo/ui/src/v2/theme.css";
```

The v2 theme wipes Tailwind's default palette (`--color-*: initial`), keeping only `transparent`, `black`, `white`, and the Radix scales below. Within a v2 build these color utilities are global — there is no `[data-theme="v2"]` ancestor requirement. A given build is either v1 or v2; the two do not coexist on the same page.

## Do / don't

### Use the numbered scale

```tsx
// Good — numbered scale step
<div className="border border-sand-7 bg-sand-3">...</div>

// Bad — hardcoded hex
<div className="border border-[#cfceca] bg-[#f1f0ef]">...</div>

// Bad — v1 semantic color names in a v2 component
<div className="border border-border-low bg-subtle">...</div>
```

### Respect step ranges

```tsx
// Good — step 3 for element background, step 11 for text
<button className="bg-sand-3 text-sand-12 hover:bg-sand-4">Save</button>

// Bad — step 11 is a text step, not a background step
<button className="bg-sand-11 text-white">Save</button>
```

### Do not mix v1 and v2 colors

```tsx
// Bad — mixing v1 (txt-primary) and v2 (sand-3) in one component
<div className="bg-sand-3 text-txt-primary">...</div>

// Good — all v2
<div className="bg-sand-3 text-sand-12">...</div>
```

### Let the theme handle dark mode

```tsx
// Bad — manual dark: overrides for v2 colors
<div className="bg-sand-1 dark:bg-sand-12">...</div>

// Good — just use the scale; dark values come from the theme scope
<div className="bg-sand-1">...</div>
```

### Solid backgrounds (steps 9–10)

Steps 9 and 10 are designed for prominent, solid-color backgrounds (primary buttons, badges, banners). Most step 9 colors are designed for white foreground text. Exceptions: **sky**, **amber** are designed for dark foreground text on steps 9–10.

```tsx
// Good — green solid button with white text
<button className="bg-green-9 text-white hover:bg-green-10">Approve</button>

// Good — amber badge with dark text (amber 9-10 are light/bright)
<span className="bg-amber-9 text-amber-12">Warning</span>
```
