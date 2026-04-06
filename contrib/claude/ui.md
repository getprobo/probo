# UI system (`@probo/ui`)

Shared React UI for Probo apps lives in the **`@probo/ui`** workspace package ([`packages/ui`](../../packages/ui)). This document describes **target** conventions for building and styling those components.

**Today’s codebase does not fully match these rules.** The tree still uses layouts like `Atoms/`, `Molecules/`, and `Layouts/`, and many files mix ad-hoc Tailwind on `className` with `tailwind-variants`. Treat this guide as the direction for new work and refactors, not as a description of the current tree.

For data loading and GraphQL on the console, see [`contrib/claude/relay.md`](relay.md).

## Package and tooling

| Item | Convention |
|------|------------|
| Package | **`@probo/ui`** — import shared components from this package in apps. |
| Styling | **Tailwind** (project uses Tailwind v4 in `packages/ui`). |
| Variants API | **`tailwind-variants`** — `import { tv } from "tailwind-variants"` to define component styles and slot class names. |

Preview components with Storybook from `packages/ui`: `npm run dev` (Storybook on port 6006 per `package.json`).

## `tailwind-variants` and `className`

In a **single component file**, do **not** mix arbitrary Tailwind utility strings on `className` with `tailwind-variants` for the same styling concerns. Put layout and look in **`tv` variants and `slots`** (and the APIs `tv` exposes for overrides). If consumers need extensibility, expose it through variant props or documented slot/class hooks—not by sprinkling raw utilities beside `tv()` output in the same file.

For **compound / multi-slot** components, define `tv` in a **dedicated module** (see [Variants file](#variants-file)) so loading-only code paths can import styles without pulling the full interactive implementation.

### Do / don’t: `tv` vs raw `className`

```tsx
// Bad — same file mixes tv() output with ad-hoc Tailwind on className (clsx shown for the anti-pattern)
import { clsx } from "clsx";
import { tv } from "tailwind-variants";

const row = tv({ base: "flex items-center gap-2" });
export function Row({ children }: { children: React.ReactNode }) {
  return <div className={clsx(row(), "rounded-md border border-border-low")}>{children}</div>;
}
```

```tsx
// Good — layout and look live in tv
import { tv } from "tailwind-variants";

const row = tv({
  base: "flex items-center gap-2 rounded-md border border-border-low",
});
export function Row({ children }: { children: React.ReactNode }) {
  return <div className={row()}>{children}</div>;
}
```

```tsx
// Good — optional styling toggles use tv variants, not extra className strings in this file
import { tv } from "tailwind-variants";

const row = tv({
  base: "flex items-center gap-2",
  variants: {
    bordered: { true: "rounded-md border border-border-low", false: "" },
  },
  defaultVariants: { bordered: true },
});
export function Row({ bordered, children }: { bordered?: boolean; children: React.ReactNode }) {
  return <div className={row({ bordered })}>{children}</div>;
}
```

## Folder layout

**Simple and layout primitives** belong in **usage-oriented** folders:

- `typography/`
- `form/`
- `layouts/`

**Other components** live in a folder **named after the component** (e.g. `ImageCard/`), with optional split files for subparts.

### Do / don’t: folder placement

```text
// Good — target layout (usage folders for primitives, component folder for composites)
packages/ui/src/
  media/Image.tsx
  media/ImageSkeleton.tsx
  typography/Text.tsx
  typography/TextSkeleton.tsx
  form/Field.tsx
  layouts/CenteredLayout.tsx
  ImageCard/variants.ts
  ImageCard/ImageCard.tsx
  ImageCard/ImageCardSkeleton.tsx

// Bad — ad-hoc placement for a simple primitive (should live under typography / form / layouts)
packages/ui/src/RandomFolder/Text.tsx
```

## Skeletons

For each meaningful component, provide a paired loading UI:

- Naming: **`ComponentName`** and **`ComponentNameSkeleton`** (e.g. `Text` / `TextSkeleton`).

A partial precedent today: [`CenteredLayoutSkeleton`](../../packages/ui/src/Layouts/CenteredLayout.tsx) alongside the layout component.

### Do / don’t: skeleton naming

```tsx
// Good — paired names
export function Text(props: TextProps) { /* … */ }
export function TextSkeleton() { /* … */ }

// Bad — unrelated name or missing pair
export function Text(props: TextProps) { /* … */ }
export function LoadingText() { /* … */ } // use TextSkeleton instead
```

## Compound modules (e.g. `ImageCard`)

Multi-region UI (card shell, media, text column, etc.) is exported as a **compound module** with static properties on one object.

### Folder and exports

- One directory per feature component (e.g. `ImageCard/`). Heavy logic may live in **separate files**; the public surface remains **`ImageCard`** with attached properties.
- **`ImageCard`** — the compound module (named/default export as established in the package).
- **`ImageCard.Root`** — top-level container **when it may hold business logic** (state, effects, data wiring, etc.).
- **`ImageCard.Shell`** — **pure layout shell**: takes **`image`** and **`text`** (and other region) **props**—each a `ReactNode`—and places them in the matching **`tv` slots**. **No children** for layout regions on `Shell`; **no state or logic** in `Shell`. If the outer wrapper is layout-only, expose it as **`Shell`**, not **`Root`**.
- **`Image`** and **`Text`** — **shared primitives** from **`@probo/ui`** (e.g. typography / media folders), not namespaced under **`ImageCard`**. **`Root`** composes them into **`Shell`**’s **`image`** / **`text`** props; apps import the same **`Image`** / **`Text`** everywhere.

**`Root` vs `Shell`:** use **`Root`** when the container owns logic; use **`Shell`** for a presentational outer frame. **`Root` may render `Shell`** inside when logic sits outside the styled layout.

### `tailwind-variants` slots

For this pattern, model regions with **`tv` `slots`** named consistently with the layout—for the example above:

- `shell`
- `image`
- `text`

Add or rename slots when the layout has more or different regions. **`Shell`** applies the matching slot output on its wrappers; **`Image`** / **`Text`** stay free of **`ImageCard`**-specific layout—keep the [no-mixing rule](#tailwind-variants-and-classname) in each file.

### Do / don’t: compound API and slots

`variants.ts` holds `tv`; **`Shell`** applies slot class names on its wrapping tags only (no duplicate Tailwind strings for those regions in the same file).

```ts
// ImageCard/variants.ts — Good
import { tv } from "tailwind-variants";

export const imageCard = tv({
  slots: {
    shell: "flex gap-4 rounded-lg border border-border-low p-4",
    image: "shrink-0 overflow-hidden rounded-md",
    text: "min-w-0 flex-1 flex flex-col gap-1",
  },
});
```

`Shell` calls **`imageCard()`** (or **`imageCard({ … })`** when the layout has variants), destructures **`shell`**, **`image`**, and **`text`**, and mounts each slot’s class name on a **wrapper element** around the prop node. **`Image`** and **`Text`** supply semantics and styling for media and copy; **`Shell`** only owns the **card layout slot wrappers**.

```tsx
// ImageCard/ImageCardShell.tsx — Good — slot class names on wrapping tags
import { imageCard } from "./variants";

export function ImageCardShell({ image, text }: { image: React.ReactNode; text: React.ReactNode }) {
  const { shell, image: imageSlot, text: textSlot } = imageCard();
  return (
    <div className={shell()}>
      <div className={imageSlot()}>{image}</div>
      <div className={textSlot()}>{text}</div>
    </div>
  );
}
```

```tsx
// ImageCard/ImageCard.tsx — Good — Root owns logic; Shell receives region nodes as props
import { Image, Text } from "@probo/ui";

function ImageCardRoot({ image, text }: { image: React.ReactNode; text: React.ReactNode }) {
  const id = useId();
  // state, effects, data wiring …
  return (
    <ImageCard.Shell
      image={<Image>{image}</Image>}
      text={<Text>{text}</Text>}
    />
  );
}

// Bad — Shell takes regions as children instead of image / text props
// <ImageCard.Shell>
//   <Image>…</Image>
//   <Text>…</Text>
// </ImageCard.Shell>

// Bad — data hooks or state live on Shell
function ImageCardShellWithData({ image, text }: { image: React.ReactNode; text: React.ReactNode }) {
  const data = useQuery(/* … */); // move to Root (or above)
  return (
    <div>
      {image}
      {text}
    </div>
  );
}
```

(The snippets above are illustrative; names and props should match the real component.)

## Skeleton placement and composition

For compound components, the **card-level skeleton is not nested** on the module as `ImageCard.Skeleton`. Export **`ImageCardSkeleton`** as a **separate top-level symbol** (e.g. `ImageCardSkeleton.tsx` or the folder barrel) so routes can depend on **loading UI + shell layout** without importing the full `ImageCard` graph—smaller initial bundles for skeleton-first views. That also avoids pulling in **Radix UI** and other dependencies that are **not needed at load time** for the skeleton-only path.

**Implementation:** `ImageCardSkeleton` should **reuse the same layout as the real card** by rendering **`ImageCard.Shell`** with the same **`image` / `text` props** as **`ImageCard.Root`**, but passing **skeleton primitives** instead of **`Image`** / **`Text`**:

- **`image`** → **`ImageSkeleton`**
- **`text`** → **`TextSkeleton`**

`Root` composes real content with **`Image`** and **`Text`** (same imports as elsewhere in the app). The skeleton passes **`ImageSkeleton`** and **`TextSkeleton`** directly into **`Shell`** so loading views avoid **`Image`** / **`Text`** when that keeps bundles or behavior simpler.

Reuse existing **`ImageSkeleton`** / **`TextSkeleton`** from typography or media primitives when available; avoid duplicate one-off pulse blocks.

### Do / don’t: skeleton imports and composition

```tsx
// Bad — skeleton nested on the compound object (pulls full card module into the route)
import { ImageCard } from "@probo/ui";
<ImageCard.Skeleton />

// Good — top-level skeleton export; reuse Shell + ImageSkeleton / TextSkeleton
import { ImageCard, ImageCardSkeleton } from "@probo/ui";

// Inside ImageCardSkeleton.tsx (conceptually):
export function ImageCardSkeleton() {
  return (
    <ImageCard.Shell
      image={<ImageSkeleton />}
      text={<TextSkeleton />}
    />
  );
}
```

The important part is **separate `ImageCardSkeleton` export**, **one `Shell` API** (`image` / `text` props), **shared shell layout**, and **reused `ImageSkeleton` / `TextSkeleton`**.

## Variants file

Keep the **`tv({ slots: { … } })` definition** (and derived slot functions) in a **standalone file**, conventionally **`variants.ts`** next to the component folder. Import it from **`Shell`** and **skeleton** modules so skeleton entry points can pull **variants + shell** without the rest of the compound module’s business logic.

### Do / don’t: colocating `tv` with the heavy module

```tsx
// Bad — variants defined only inside ImageCard.tsx; ImageCardSkeleton imports ImageCard and drags Root / hooks
// ImageCard.tsx
const imageCard = tv({ slots: { shell: "...", image: "...", text: "..." } });

// Good — shared variants module imported by Shell and ImageCardSkeleton only
// variants.ts — export imageCard (or slot helpers)
// ImageCardShell.tsx — import { imageCard } from "./variants"
// ImageCardSkeleton.tsx — import { imageCard } from "./variants"
```
