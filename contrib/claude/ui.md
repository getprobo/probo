# UI system (`@probo/ui` v2 kit)

Shared React UI for Probo apps lives in the **`@probo/ui`** workspace package ([`packages/ui`](../../packages/ui)). The **v2 kit** ([`packages/ui/src/v2`](../../packages/ui/src/v2)) is the target system: a flat set of components styled on top of a headless primitive library, consuming the Radix-scale v2 theme. This document describes how to build and style those components.

These rules are the **source of truth**. The legacy tree (`Atoms/`, `Molecules/`, `Layouts/`, `clsx`-mixed `className`, imperative `DialogRef`) is non-compliant code to migrate, not precedent.

## Related guides

| Topic | Guide |
|-------|--------|
| Component shape, props, naming/suffixes | [`contrib/claude/react-components.md`](react-components.md) |
| v2 design tokens (color, type, radius, shadow, spacing) | [`contrib/claude/v2-tokens.md`](v2-tokens.md) |
| App folder layout and special folders | [`contrib/claude/app-arborescence.md`](app-arborescence.md) |
| Error boundaries and error/fallback props | [`contrib/claude/error-handling.md`](error-handling.md) |
| Relay data loading | [`contrib/claude/relay.md`](relay.md) |

## Package and tooling

| Item | Convention |
|------|------------|
| Package | **`@probo/ui`** — v2 components under `src/v2`. Apps opt into v2 by importing the v2 theme (see [`v2-tokens.md`](v2-tokens.md)). |
| Styling | **Tailwind v4** with the Radix-scale tokens (`bg-sand-3`, `text-sand-12`, `rounded-3`, `text-4`, …). |
| Variants API | **`tailwind-variants`** only — `import { tv } from "tailwind-variants"`. |
| Class composition | **Do not use `clsx` or `tailwind-merge`.** All conditional styling goes through `tv` variants and slots. |
| Headless primitives | **Base UI** (`@base-ui-components/react`). We **style** these primitives; we do not re-implement their behavior. |

Preview components with Storybook from `packages/ui`: `npm run dev` (Storybook on port 6006).

> Base UI is migrating its package name from `@base-ui-components/react` to `@base-ui/react`. Import from whichever name the installed version publishes; examples below use `@base-ui-components/react`.

## Headless primitives: style, don't re-implement

Interactive components (dialogs, popovers, menus, selects, tabs, tooltips) are **Base UI primitives with our styling applied** — nothing more. The job of a v2 component is to bind `tv` classes to the primitive's parts. Do **not** add a custom behavior layer on top.

Rules:

- **Use the primitive's controlled API as-is.** A dialog is controlled with `open` / `onOpenChange` — the exact same API Base UI exposes. Opening *our* dialog is opening *the lib's* dialog.
- **No custom imperative ref API.** Never invent `useDialogRef()` / `ref.current.open()` / `ref.current.close()`. If imperative control is genuinely needed, use the primitive's own mechanism (e.g. Base UI's `Dialog.createHandle()` / `actionsRef`), never a hand-rolled `useRef` + `useEffect` shim.
- **No `cloneElement` / `Children.map` plumbing.** Compose with the primitive's parts and `asChild`-style props the library provides, not by cloning children to inject className/handlers.
- **No local mirror state.** Don't copy `open` into `useState` and sync it with `useEffect`; pass `open`/`onOpenChange` straight through, or let the primitive stay uncontrolled.

### Do / don't: dialog wrapper

```tsx
// Bad — hand-rolled imperative ref, mirrored open state, cloneElement plumbing
export const useDialogRef = () => useRef(null);

export function Dialog({ trigger, ref, children }: Props) {
  const [open, setOpen] = useState(false);
  useEffect(() => {
    if (ref) ref.current = { open: () => setOpen(true), close: () => setOpen(false) };
  });
  // ... Children.map / cloneElement to inject classes ...
  return <Root open={open} onOpenChange={setOpen}>{/* … */}</Root>;
}
```

```tsx
// Good — thin styling over Base UI; consumers use the lib's open/onOpenChange directly
import { Dialog as BaseDialog } from "@base-ui-components/react/dialog";
import { tv } from "tailwind-variants";

const dialog = tv({
  slots: {
    backdrop: "fixed inset-0 bg-sand-12/40",
    popup: "fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-4 bg-sand-2 p-6 shadow-4",
    title: "text-4 font-medium text-sand-12",
  },
});

export type DialogProps = ComponentProps<typeof BaseDialog.Root>;

export function Dialog(props: DialogProps) {
  return <BaseDialog.Root {...props} />;
}

export function DialogPopup({ children, ...props }: ComponentProps<typeof BaseDialog.Popup>) {
  const { backdrop, popup } = dialog();
  return (
    <BaseDialog.Portal>
      <BaseDialog.Backdrop className={backdrop()} />
      <BaseDialog.Popup className={popup()} {...props}>
        {children}
      </BaseDialog.Popup>
    </BaseDialog.Portal>
  );
}
```

```tsx
// Good — consumer controls it the same way they'd control the Base UI dialog
const [open, setOpen] = useState(false);

<Dialog open={open} onOpenChange={setOpen}>
  <DialogTrigger>Open</DialogTrigger>
  <DialogPopup>
    <DialogTitle>Delete third party</DialogTitle>
    {/* … */}
  </DialogPopup>
</Dialog>
```

## `tailwind-variants` and `className`

In a **single component file**, do **not** mix arbitrary Tailwind utility strings on `className` with `tailwind-variants` for the same styling concern. Layout and look live in **`tv` slots and variants** (and the override APIs `tv` exposes). Extensibility is exposed through variant props or documented slot/class hooks — never by sprinkling raw utilities (or `clsx`) beside `tv()` output.

For **compound / multi-slot** components, define `tv` in a **dedicated `variants.ts` module** (see [Variants file](#variants-file)) so loading-only code paths can import styles without pulling the full interactive implementation.

### Do / don't: `tv` vs raw `className`

```tsx
// Bad — same file mixes tv() output with ad-hoc Tailwind / clsx on className
import { clsx } from "clsx";
import { tv } from "tailwind-variants";

const row = tv({ base: "flex items-center gap-2" });
export function Row({ children }: { children: ReactNode }) {
  return <div className={clsx(row(), "rounded-3 border border-sand-6")}>{children}</div>;
}
```

```tsx
// Good — layout and look live in tv
import { tv } from "tailwind-variants";

const row = tv({
  base: "flex items-center gap-2 rounded-3 border border-sand-6",
});
export function Row({ children }: { children: ReactNode }) {
  return <div className={row()}>{children}</div>;
}
```

```tsx
// Good — optional styling toggles use tv variants, not extra className strings
import { tv } from "tailwind-variants";

const row = tv({
  base: "flex items-center gap-2",
  variants: {
    bordered: { true: "rounded-3 border border-sand-6", false: "" },
  },
  defaultVariants: { bordered: true },
});
export function Row({ bordered, children }: { bordered?: boolean; children: ReactNode }) {
  return <div className={row({ bordered })}>{children}</div>;
}
```

## No structure-changing variants

Variants tune **look** (size, tone, density) — they must not change a component's **structure, semantics, or prop contract**. When a "variant" would render a different element, accept different props, or fork the behavior, build a **separate component** instead. This keeps each component's typing simple and its rendered element predictable.

The clearest case is the button family: a clickable action, a styled `<a>`, and a router link are three components, not one `Button` with an `as`/`href`/`to` union.

### Do / don't: separate components over polymorphic props

```tsx
// Bad — one component forks structure on props; typing becomes a union mess
type ButtonProps =
  | { as?: "button"; onClick: () => void }
  | { as: "a"; href: string }
  | { as: "link"; to: string };

export function Button(props: ButtonProps) {
  if (props.as === "a") return <a href={props.href} className={button()} />;
  if (props.as === "link") return <RouterLink to={props.to} className={button()} />;
  return <button onClick={props.onClick} className={button()} />;
}
```

```tsx
// Good — three flat components sharing the same tv styles
// variants.ts
export const button = tv({ base: "inline-flex items-center …", variants: { /* size, tone */ } });

// Button.tsx — renders <button>
export function Button(props: ComponentProps<"button">) {
  return <button className={button()} {...props} />;
}

// Anchor.tsx — renders <a>
export function Anchor(props: ComponentProps<"a">) {
  return <a className={button()} {...props} />;
}

// Link.tsx — renders a router link
export function Link(props: ComponentProps<typeof RouterLink>) {
  return <RouterLink className={button()} {...props} />;
}
```

Size/tone differences (`size="sm"`, `tone="danger"`) are legitimate `tv` variants — they don't change the element or props.

## Props typing

When a component renders a native HTML element (or a single primitive part) as its top-level node, **merge the component's own props with that element's intrinsic props** via `ComponentProps`. Destructure custom props and spread the rest so callers can pass standard attributes (`id`, `className`, `aria-*`, event handlers) without wrapper boilerplate.

### Do / don't: props merging

```tsx
// Good — own props merged with the native element's props, rest spread onto <span>
type TextProps = ComponentProps<"span"> & { tone?: "default" | "muted" };

export function Text(props: TextProps) {
  const { tone = "default", className, ...spanProps } = props;
  return <span className={text({ tone, className })} {...spanProps} />;
}
```

```tsx
// Bad — only custom props accepted; callers cannot set id, className, aria-*, etc.
type TextProps = { children: ReactNode };

export function Text(props: TextProps) {
  return <span>{props.children}</span>;
}
```

## Icons

Icons come from two sources, in this order of preference:

1. **`@phosphor-icons/react`** — the default icon library. Import the specific icon directly: `import { CookieIcon } from "@phosphor-icons/react"`. Prefer phosphor whenever it has the icon you need.
2. **`@probo/ui` `Icon*` set** — curated in-house icons. Use these only when phosphor has no suitable equivalent or you need a bespoke Probo-branded icon.

**Never use emoji characters (🍪, ✅, ⚠️, …) as icons.** Emojis render inconsistently, don't inherit `currentColor`, and can't be sized like an SVG. If neither source has what you need, add the icon to `@probo/ui`.

### Phosphor import style

Always import phosphor icons by their **`Icon`-suffixed name** (e.g. `EyeIcon`, `CookieIcon`). **Never** import the bare name and alias it with an `Icon` prefix.

```tsx
// Bad — bare name aliased to add an Icon prefix
import { Eye as IconEye } from "@phosphor-icons/react";

// Good — use the Icon-suffixed export directly
import { EyeIcon, EyeSlashIcon } from "@phosphor-icons/react";
```

### Do / don't: icon source

```tsx
// Bad — emoji used as an icon
<div className="mb-2 text-9">🍪</div>
```

```tsx
// Good — phosphor icon as the default choice
import { CookieIcon } from "@phosphor-icons/react";

<CookieIcon size={48} weight="duotone" className="text-sand-11" />
```

## Folder layout

The v2 tree is **flat** — there is no `Atoms/` / `Molecules/` / `Layouts/` hierarchy.

- **Simple and layout primitives** live in **usage-oriented** folders: `typography/`, `form/`, `layouts/`.
- A **complex component gets its own folder** named after the component (e.g. `Dropdown/`, `Dialog/`, `ImageCard/`), holding its parts, `variants.ts`, and skeleton.

### Do / don't: folder placement

```text
// Good — usage folders for primitives, component folder for composites
packages/ui/src/v2/
  typography/Text.tsx
  typography/TextSkeleton.tsx
  form/Field.tsx
  layouts/CenteredLayout.tsx
  Dialog/Dialog.tsx
  Dialog/DialogPopup.tsx
  Dialog/variants.ts
  Dropdown/Dropdown.tsx
  Dropdown/DropdownItem.tsx
  ImageCard/variants.ts
  ImageCard/ImageCardRoot.tsx
  ImageCard/ImageCardShell.tsx
  ImageCard/ImageCardSkeleton.tsx

// Bad — primitive buried in an ad-hoc folder (belongs under typography/)
packages/ui/src/v2/RandomFolder/Text.tsx

// Bad — legacy classification folders
packages/ui/src/v2/Atoms/Button.tsx
```

### Naming

UI-kit components use **bare names** (no role suffix): `Button`, `Anchor`, `Link`, `Text`, `List`, `ListItem`, `Dialog`. Parts of a complex component are prefixed with the component name (`DialogPopup`, `DropdownItem`). App-level components use the suffix taxonomy in [`react-components.md`](react-components.md#naming-and-suffixes).

## Primitives vs compound components

Components fall into two categories: **primitives** and **compound** components.

### Primitives

**Primitives** (`Text`, `Image`, form inputs, layout helpers, `ListItem`) are self-contained — they render a single semantic element with their own styling. A primitive **is its own shell**: there is no separate shell wrapper. Each primitive has a paired skeleton (`TextSkeleton`, `ImageSkeleton`) that matches its dimensions.

### Compound components

**Compound components** (`ImageCard`, …) assemble multiple primitives into a larger region. When logic (state, effects, data) lives inside the top-level component, a **shell** separates layout from behavior:

- **Shell** — pure layout frame that accepts region props (`image`, `text`, …) as `ReactNode` and applies `tv` slot classes. No state, no effects, no data.
- **Root** — owns the logic and renders the shell, passing primitives into its region props.
- **Skeleton** — reuses the **same shell** with skeleton primitives, so the loading placeholder is structurally identical without pulling in the logic graph.

If a compound component is purely presentational (no logic), there is no Root — expose only the Shell.

## Skeletons

Every meaningful component provides a paired loading UI named `ComponentName` / `ComponentNameSkeleton` (e.g. `Text` / `TextSkeleton`).

Skeletons are **typography and shapes only** — pulse blocks sized to match the real layout (a `TextSkeleton` matches a line of text; a `Dialog` exposes a `DialogSkeleton` matching its frame). They must render instantly and carry **no data-fetching logic**.

### Skeletons must stay out of the heavy bundle

A skeleton's whole point is to render *before* the real component (and its dependencies) load. A skeleton must therefore be importable **without dragging in Base UI or other heavy interactive dependencies**.

- Keep `tv` slot definitions in a standalone **`variants.ts`** (see [Variants file](#variants-file)). The shell and the skeleton import `variants.ts`; neither imports the Root's logic.
- Export each `*Skeleton` as a **standalone named export** from its own module — never as a property on a namespace object (`Dialog.Skeleton`) and never re-exported from a barrel that also pulls the interactive implementation into the same chunk.
- A complex component **exposes its own skeleton** (`DialogSkeleton`, `ImageCardSkeleton`) so pages can show a faithful placeholder; that skeleton renders the **shell + skeleton primitives**, importing none of the Base UI parts.

### Do / don't: skeleton naming and bundle safety

```tsx
// Good — paired names, skeleton imports only shell + skeleton primitives
export function ImageCard(props: ImageCardProps) { /* … */ }

// ImageCardSkeleton.tsx — no Base UI / Root imports reach this module
import { ImageCardShell } from "./ImageCardShell";
import { ImageSkeleton } from "../media/ImageSkeleton";
import { TextSkeleton } from "../typography/TextSkeleton";

export function ImageCardSkeleton() {
  return <ImageCardShell image={<ImageSkeleton />} text={<TextSkeleton />} />;
}
```

```tsx
// Bad — skeleton nested on a namespace object (pulls the full interactive module in)
import { ImageCard } from "@probo/ui";
<ImageCard.Skeleton />

// Bad — unrelated name / missing pair
export function LoadingText() { /* … */ } // use TextSkeleton instead
```

## Compound component structure (e.g. `ImageCard`)

Multi-region UI is exported as **individual named exports** — one per sub-component — all prefixed with the feature name (e.g. `ImageCardRoot`, `ImageCardShell`, `ImageCardSkeleton`). **Do not** group sub-components as static properties on a namespace object (`ImageCard.Root`, …); flat named exports enable proper tree shaking and keep heavy dependencies out of loading-time bundles.

- One directory per feature component. Heavy logic may live in separate files; each public part is a standalone named export.
- **`ImageCardRoot`** — top-level container **when it holds logic** (state, effects, data wiring).
- **`ImageCardShell`** — **pure layout shell**: takes region props (`image`, `text`, …), each a `ReactNode`, and places them in matching `tv` slots. No children for layout regions, no state, no logic. If the outer wrapper is layout-only, expose it as `ImageCardShell`, not `ImageCardRoot`.
- **`Image`** and **`Text`** — shared primitives from the kit, not prefixed under `ImageCard`. `ImageCardRoot` composes them into `ImageCardShell`'s region props.

### `tailwind-variants` slots

Model regions with `tv` `slots` named after the layout:

```ts
// ImageCard/variants.ts
import { tv } from "tailwind-variants";

export const imageCard = tv({
  slots: {
    shell: "flex gap-4 rounded-4 border border-sand-6 p-4",
    image: "shrink-0 overflow-hidden rounded-3",
    text: "min-w-0 flex-1 flex flex-col gap-1",
  },
});
```

`ImageCardShell` calls `imageCard()`, destructures the slots, and mounts each slot's class on a wrapper element around the prop node:

```tsx
// ImageCard/ImageCardShell.tsx — slot classes on wrapping tags
import { imageCard } from "./variants";

export function ImageCardShell({ image, text }: { image: ReactNode; text: ReactNode }) {
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
// ImageCard/ImageCardRoot.tsx — Root owns logic; Shell receives region nodes as props
import { Image, Text } from "@probo/ui";
import { ImageCardShell } from "./ImageCardShell";

export function ImageCardRoot({ image, text }: { image: ReactNode; text: ReactNode }) {
  // state, effects, data wiring …
  return <ImageCardShell image={<Image>{image}</Image>} text={<Text>{text}</Text>} />;
}

// Bad — Shell takes regions as children instead of image / text props
// Bad — data hooks or state live on Shell (move to Root or above)
```

## Variants file

Keep the `tv({ slots: { … } })` definition (and derived slot functions) in a standalone **`variants.ts`** next to the component folder. Import it from the shell and skeleton modules so skeleton entry points can pull **variants + shell** without the rest of the compound component's business logic (and without Base UI).

```tsx
// Bad — variants defined inside ImageCardRoot.tsx; the skeleton importing it drags Root + hooks (+ Base UI)
// ImageCardRoot.tsx
const imageCard = tv({ slots: { shell: "...", image: "...", text: "..." } });

// Good — shared variants module imported by ImageCardShell and ImageCardSkeleton only
// variants.ts        — export imageCard (or slot helpers)
// ImageCardShell.tsx — import { imageCard } from "./variants"
// ImageCardSkeleton.tsx — import { imageCard } from "./variants"
```

## User feedback (toasts)

Transient feedback for an action's outcome uses **Base UI's Toast** (`@base-ui-components/react/toast`) — never `alert`, a hand-rolled banner, or a `console.log`. As with every other primitive, we **style** Base UI's toast; we do not build our own toast system. The legacy kit `useToast` / `Toaster` is non-compliant and is being removed — do not use it in v2.

The kit exposes a styled **`Toaster`** (a `Toast.Portal` + `Toast.Viewport` rendering styled `Toast.Root`s, keyed off each toast's `type`). Mount Base UI's `Toast.Provider` and the `Toaster` **once** at the app root; everything else queues toasts through Base UI's manager.

```tsx
// app root — Base UI provider + the kit's styled viewport, mounted once
import { Toast } from "@base-ui-components/react/toast";
import { Toaster } from "@probo/ui";

<Toast.Provider>
  <App />
  <Toaster />
</Toast.Provider>
```

Queue a toast with `Toast.useToastManager().add(...)` — the same API Base UI exposes. Use `type` to drive the styled variant.

```tsx
import { Toast } from "@base-ui-components/react/toast";

function CreateMeasureButton() {
  const toast = Toast.useToastManager();
  const { t } = useTranslation();
  const [createMeasure] = useMutation<CreateMeasureMutation>(createMeasureMutation);

  function onCreate() {
    createMeasure({
      variables: { input, connections: [connectionId] },
      onCompleted() {
        toast.add({ title: t("measures.created"), type: "success" });
      },
      onError(error) {
        toast.add({
          title: t("common.error"),
          description: formatError(t("measures.createFailed"), error as GraphQLError),
          type: "error",
        });
      },
    });
  }
  // …
}
```

For code outside the React tree, create a global manager with `Toast.createToastManager()` and pass it to `Toast.Provider` via `toastManager` — still the same renderer.

Choose **toast vs. inline** by where the message belongs:

- **Toast** — the result of an action not tied to a specific field: a successful save, a delete, an unexpected mutation failure.
- **Inline** — validation tied to a field or region: render it in `Field.Error` (see [`forms.md`](forms.md)) or a section's error UI (see [`error-handling.md`](error-handling.md)), not a toast.

```tsx
// Bad — the legacy kit hook (removed in v2)
const { toast } = useToast();
toast({ title: "Saved", variant: "success" });

// Bad — field validation surfaced as a toast (belongs inline on the field)
toast.add({ title: "Name is required", type: "error" });

// Bad — browser alert / ad-hoc UI for feedback
alert("Saved!");
```

## Empty states

A `*List` (or any collection region) renders an **empty state** when it has no items — never a blank gap. Empty states are part of the component, not an afterthought, and follow the `*Empty` suffix when extracted (see [`react-components.md`](react-components.md#naming-and-suffixes)).

An empty state has: an icon (phosphor — never emoji), a short heading, optional one-line guidance, and, when the user can act, the primary call to action (gated by permission — see [`permissions.md`](permissions.md)).

```tsx
// Good — collection renders its own empty state
{measures.length === 0
  ? <MeasuresEmpty canCreate={canCreate} />
  : measures.map((m) => <MeasureListItem key={m.id} measureKey={m} />)}
```

Distinguish empty (no data) from loading (`*Skeleton`) from error (`*Error`) — they are three different states, not one.

## Accessibility

Base UI primitives ship correct roles, focus management, and keyboard interaction — **do not re-implement or override them.** Our job is to keep that behavior intact while styling:

- Keep accessible labels: every control has a visible label or an `aria-label`; icon-only buttons (`Button icon={…}`) require an `aria-label`.
- Don't strip `aria-*` / `role` that primitives set, and don't trap or override focus the primitive manages.
- Convey state with more than color (e.g. an icon + text alongside a `red-*` tone), so meaning survives for color-blind users — the [token contrast guarantees](v2-tokens.md#contrast-guarantees) cover text legibility, not state encoding.
- Use semantic elements (`<button>`, `<a>`, `<nav>`, headings) — see the [Button vs Anchor vs Link](#no-structure-changing-variants) split.
