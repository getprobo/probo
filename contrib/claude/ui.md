# UI system (`@probo/ui`)

Shared React UI for Probo apps lives in the **`@probo/ui`** workspace package ([`packages/ui`](../../packages/ui)). This document describes **target** conventions for building and styling those components.

**Today's codebase does not fully match these rules.** The tree still uses layouts like `Atoms/`, `Molecules/`, and `Layouts/`, and many files mix ad-hoc Tailwind on `className` with `tailwind-variants`. Treat this guide as the direction for new work and refactors, not as a description of the current tree.

For data loading and GraphQL on the console, see [`contrib/claude/relay.md`](relay.md).

## Package and tooling

| Item | Convention |
|------|------------|
| Package | **`@probo/ui`** — import shared components from this package in apps. |
| Styling | **Tailwind** (project uses Tailwind v4 in `packages/ui`). |
| Variants API | **`tailwind-variants`** — `import { tv } from "tailwind-variants"` to define component styles and slot class names. |

Preview components with Storybook from `packages/ui`: `npm run dev` (Storybook on port 6006 per `package.json`).

## Props typing

When a component renders a native HTML element as its top-level node (not a custom component), **merge the component's own props with that element's intrinsic props** via `ComponentProps`. Destructure custom props and spread the rest onto the element so callers can pass standard HTML attributes (`id`, `className`, `aria-*`, event handlers, etc.) without wrapper boilerplate.

### Do / don't: props merging

```tsx
// Good — own props merged with the native element's props, rest spread onto <span>
type MyProps = ComponentProps<"span"> & { myPropName: string };

export function MyComponent(props: MyProps) {
  const { myPropName, ...spanProps } = props;

  return <span {...spanProps}>{myPropName}</span>;
}
```

```tsx
// Bad — only custom props accepted; callers cannot set id, className, aria-*, etc.
type MyProps = { myPropName: string };

export function MyComponent(props: MyProps) {
  return <span>{props.myPropName}</span>;
}
```

## `tailwind-variants` and `className`

In a **single component file**, do **not** mix arbitrary Tailwind utility strings on `className` with `tailwind-variants` for the same styling concerns. Put layout and look in **`tv` variants and `slots`** (and the APIs `tv` exposes for overrides). If consumers need extensibility, expose it through variant props or documented slot/class hooks—not by sprinkling raw utilities beside `tv()` output in the same file.

For **compound / multi-slot** components, define `tv` in a **dedicated module** (see [Variants file](#variants-file)) so loading-only code paths can import styles without pulling the full interactive implementation.

### Do / don't: `tv` vs raw `className`

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

### Do / don't: folder placement

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
  ImageCard/ImageCardRoot.tsx
  ImageCard/ImageCardShell.tsx
  ImageCard/ImageCardSkeleton.tsx

// Bad — ad-hoc placement for a simple primitive (should live under typography / form / layouts)
packages/ui/src/RandomFolder/Text.tsx
```

## Primitives vs compound components

Components in `@probo/ui` fall into two categories: **primitives** and **compound** components.

### Primitives

**Primitives** (`Text`, `Image`, form inputs, layout helpers) are self-contained — they render a single semantic element with its own styling. A primitive **is its own shell**: it owns both its layout footprint and its visual output, so there is no separate shell wrapper. Each primitive has a paired skeleton (`TextSkeleton`, `ImageSkeleton`) that matches its dimensions.

### Compound components

**Compound components** (`ImageCard`, …) assemble multiple primitives into a larger UI region. When logic (state, effects, data fetching) lives inside the top-level component, a **shell** is required to separate layout from behavior:

- **Shell** — pure layout frame that accepts region props (`image`, `text`, …) as `ReactNode` and applies `tv` slot class names. No state, no effects, no data.
- **Root** — owns the logic and renders the shell, passing primitives into its region props.
- **Skeleton** — reuses the **same shell** with skeleton primitives, so the loading placeholder is structurally identical to the real component without pulling in the logic graph.

The shell exists so that **skeletons can share the exact same layout** as the real component without importing Root and its dependencies. If the compound component is **purely presentational** (no logic needed), there is no Root — expose only the Shell.

## Skeletons

For each meaningful component, provide a paired loading UI:

- Naming: **`ComponentName`** and **`ComponentNameSkeleton`** (e.g. `Text` / `TextSkeleton`).

A partial precedent today: [`CenteredLayoutSkeleton`](../../packages/ui/src/Layouts/CenteredLayout.tsx) alongside the layout component.

### Do / don't: skeleton naming

```tsx
// Good — paired names
export function Text(props: TextProps) { /* … */ }
export function TextSkeleton() { /* … */ }

// Bad — unrelated name or missing pair
export function Text(props: TextProps) { /* … */ }
export function LoadingText() { /* … */ } // use TextSkeleton instead
```

## Compound component structure (e.g. `ImageCard`)

Multi-region UI (card shell, media, text column, etc.) is exported as **individual named exports** — one per sub-component — all prefixed with the feature name (e.g. `ImageCardRoot`, `ImageCardShell`, `ImageCardSkeleton`). **Do not** group sub-components as static properties on a single namespace object (`ImageCard.Root`, `ImageCard.Shell`, …); flat named exports enable proper tree shaking and keep unwanted third-party dependencies out of loading-time bundles.

### Folder and exports

- One directory per feature component (e.g. `ImageCard/`). Heavy logic may live in **separate files**; each public part is a **standalone named export**.
- **`ImageCardRoot`** — top-level container **when it may hold business logic** (state, effects, data wiring, etc.).
- **`ImageCardShell`** — **pure layout shell**: takes **`image`** and **`text`** (and other region) **props**—each a `ReactNode`—and places them in the matching **`tv` slots**. **No children** for layout regions on the shell; **no state or logic** in the shell. If the outer wrapper is layout-only, expose it as **`ImageCardShell`**, not **`ImageCardRoot`**.
- **`Image`** and **`Text`** — **shared primitives** from **`@probo/ui`** (e.g. typography / media folders), not prefixed under `ImageCard`. **`ImageCardRoot`** composes them into **`ImageCardShell`**'s **`image`** / **`text`** props; apps import the same **`Image`** / **`Text`** everywhere.

**Root vs Shell:** use **`ImageCardRoot`** when the container owns logic; use **`ImageCardShell`** for a presentational outer frame. **`ImageCardRoot` may render `ImageCardShell`** inside when logic sits outside the styled layout.

### `tailwind-variants` slots

For this pattern, model regions with **`tv` `slots`** named consistently with the layout—for the example above:

- `shell`
- `image`
- `text`

Add or rename slots when the layout has more or different regions. **`ImageCardShell`** applies the matching slot output on its wrappers; **`Image`** / **`Text`** stay free of **`ImageCard`**-specific layout—keep the [no-mixing rule](#tailwind-variants-and-classname) in each file.

### Do / don't: compound API and slots

`variants.ts` holds `tv`; **`ImageCardShell`** applies slot class names on its wrapping tags only (no duplicate Tailwind strings for those regions in the same file).

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

**`ImageCardShell`** calls **`imageCard()`** (or **`imageCard({ … })`** when the layout has variants), destructures **`shell`**, **`image`**, and **`text`**, and mounts each slot's class name on a **wrapper element** around the prop node. **`Image`** and **`Text`** supply semantics and styling for media and copy; **`ImageCardShell`** only owns the **card layout slot wrappers**.

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
// ImageCard/ImageCardRoot.tsx — Good — Root owns logic; Shell receives region nodes as props
import { Image, Text } from "@probo/ui";
import { ImageCardShell } from "./ImageCardShell";

function ImageCardRoot({ image, text }: { image: React.ReactNode; text: React.ReactNode }) {
  const id = useId();
  // state, effects, data wiring …
  return (
    <ImageCardShell
      image={<Image>{image}</Image>}
      text={<Text>{text}</Text>}
    />
  );
}

// Bad — Shell takes regions as children instead of image / text props
// <ImageCardShell>
//   <Image>…</Image>
//   <Text>…</Text>
// </ImageCardShell>

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

For compound components, export **`ImageCardSkeleton`** as a **separate named export** (e.g. `ImageCardSkeleton.tsx` or the folder barrel) so routes can depend on **loading UI + shell layout** without importing the full `ImageCardRoot` graph—smaller initial bundles for skeleton-first views. That also avoids pulling in **Radix UI** and other dependencies that are **not needed at load time** for the skeleton-only path.

**Implementation:** `ImageCardSkeleton` should **reuse the same layout as the real card** by rendering **`ImageCardShell`** with the same **`image` / `text` props** as **`ImageCardRoot`**, but passing **skeleton primitives** instead of **`Image`** / **`Text`**:

- **`image`** → **`ImageSkeleton`**
- **`text`** → **`TextSkeleton`**

**`ImageCardRoot`** composes real content with **`Image`** and **`Text`** (same imports as elsewhere in the app). The skeleton passes **`ImageSkeleton`** and **`TextSkeleton`** directly into **`ImageCardShell`** so loading views avoid **`Image`** / **`Text`** when that keeps bundles or behavior simpler.

Reuse existing **`ImageSkeleton`** / **`TextSkeleton`** from typography or media primitives when available; avoid duplicate one-off pulse blocks.

### Do / don't: skeleton imports and composition

```tsx
// Bad — skeleton nested on a namespace object (pulls full card module into the route)
import { ImageCard } from "@probo/ui";
<ImageCard.Skeleton />

// Good — each sub-component is a standalone named export
import { ImageCardShell, ImageCardSkeleton } from "@probo/ui";

// Inside ImageCardSkeleton.tsx (conceptually):
export function ImageCardSkeleton() {
  return (
    <ImageCardShell
      image={<ImageSkeleton />}
      text={<TextSkeleton />}
    />
  );
}
```

The important part is **separate `ImageCardSkeleton` export**, **one `ImageCardShell` API** (`image` / `text` props), **shared shell layout**, and **reused `ImageSkeleton` / `TextSkeleton`**.

## Variants file

Keep the **`tv({ slots: { … } })` definition** (and derived slot functions) in a **standalone file**, conventionally **`variants.ts`** next to the component folder. Import it from **`ImageCardShell`** and **skeleton** modules so skeleton entry points can pull **variants + shell** without the rest of the compound component's business logic.

### Do / don't: colocating `tv` with the heavy module

```tsx
// Bad — variants defined only inside ImageCardRoot.tsx; ImageCardSkeleton imports it and drags Root / hooks
// ImageCardRoot.tsx
const imageCard = tv({ slots: { shell: "...", image: "...", text: "..." } });

// Good — shared variants module imported by ImageCardShell and ImageCardSkeleton only
// variants.ts — export imageCard (or slot helpers)
// ImageCardShell.tsx — import { imageCard } from "./variants"
// ImageCardSkeleton.tsx — import { imageCard } from "./variants"
```
