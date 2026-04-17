# @probo/cookie-banner

A lightweight, GDPR-compliant cookie consent banner for the web. Works with any framework or plain HTML — zero dependencies, powered by Web Components.

## Installation

```bash
npm install @probo/cookie-banner
```

## Quick Start (Script Tag)

The fastest way to add a cookie banner to any website. Drop a single `<script>` tag into your HTML — no bundler required:

```html
<script
  src="https://unpkg.com/@probo/cookie-banner/dist/cookie-banner.iife.js"
  data-banner-id="YOUR_BANNER_ID"
  data-base-url="BASE_URL"
  data-position="bottom-left"
></script>
```

This automatically renders a fully styled consent dialog and a floating settings button so visitors can change their preferences at any time.

| Attribute        | Required | Description                                                       |
| ---------------- | -------- | ----------------------------------------------------------------- |
| `data-banner-id` | Yes      | Your banner ID from the Probo dashboard                           |
| `data-base-url`  | Yes      | The Probo cookie-banner API base URL                              |
| `data-position`  | No       | Settings button position: `bottom-left` (default), `bottom-right` |

## Themed Banner (ES Module)

Import the pre-built themed banner as an ES module for use in bundled applications (React, Vue, Svelte, etc.):

```js
import { registerThemedBanner } from "@probo/cookie-banner/themed-banner";

registerThemedBanner();
```

Then use the `<probo-cookie-banner>` custom element anywhere in your HTML or templates:

```html
<probo-cookie-banner
  banner-id="YOUR_BANNER_ID"
  base-url="BASE_URL"
  position="bottom-left"
></probo-cookie-banner>
```

### Theming with CSS Custom Properties

The themed banner renders inside a Shadow DOM and exposes CSS custom properties for styling:

```css
probo-cookie-banner {
  --probo-font-family: "Inter", sans-serif;
  --probo-bg: #ffffff;
  --probo-text: #1a1a1a;
  --probo-text-secondary: #555555;
  --probo-border: #e0e0e0;
  --probo-radius: 12px;
  --probo-shadow: 0 4px 24px rgba(0, 0, 0, 0.12);
  --probo-accent: #1a1a1a;
  --probo-accent-text: #ffffff;
  --probo-overlay: rgba(0, 0, 0, 0.4);
  --probo-z-index: 2147483646;
  --probo-btn-radius: 8px;
  --probo-font-size: 14px;
}
```

Dark mode example:

```css
@media (prefers-color-scheme: dark) {
  probo-cookie-banner {
    --probo-bg: #1a1a1a;
    --probo-text: #f0f0f0;
    --probo-text-secondary: #a0a0a0;
    --probo-border: #333333;
    --probo-accent: #f0f0f0;
    --probo-accent-text: #1a1a1a;
    --probo-overlay: rgba(0, 0, 0, 0.6);
  }
}
```

## Headless Components (Full Control)

For complete control over the UI, import the headless Web Components directly. This entrypoint gives you unstyled building blocks that you compose and style yourself:

```js
import { registerComponents } from "@probo/cookie-banner";

registerComponents();
```

Then build your own banner layout using the provided custom elements:

```html
<probo-cookie-banner-root banner-id="YOUR_BANNER_ID" base-url="BASE_URL">
  <!-- Shown when no consent is recorded -->
  <probo-banner>
    <div class="my-banner">
      <p>We use cookies to improve your experience.</p>
      <probo-accept-button>
        <button>Accept all</button>
      </probo-accept-button>
      <probo-reject-button>
        <button>Reject all</button>
      </probo-reject-button>
      <probo-customize-button>
        <button>Customize</button>
      </probo-customize-button>
    </div>
  </probo-banner>

  <!-- Shown when visitor clicks "Customize" -->
  <probo-preference-panel>
    <div class="my-preferences">
      <probo-category-list>
        <template>
          <div class="category">
            <span data-slot="name"></span>
            <span data-slot="description"></span>
            <probo-category-toggle>
              <input type="checkbox" />
            </probo-category-toggle>
          </div>
          <probo-cookie-list>
            <template>
              <div class="cookie">
                <span data-slot="name"></span>
                <span data-slot="duration"></span>
              </div>
            </template>
          </probo-cookie-list>
        </template>
      </probo-category-list>
      <probo-save-button>
        <button>Save preferences</button>
      </probo-save-button>
    </div>
  </probo-preference-panel>

  <!-- Floating button to re-open preferences after consent -->
  <probo-settings-button position="bottom-left"></probo-settings-button>
</probo-cookie-banner-root>
```

### Available Components

| Component                    | Description                                                                                                  |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------ |
| `<probo-cookie-banner-root>` | Root element. Requires `banner-id` and `base-url` attributes. Manages client lifecycle and state.            |
| `<probo-banner>`             | Container shown when consent has not been given yet.                                                         |
| `<probo-accept-button>`      | Wraps a button that records "accept all" consent.                                                            |
| `<probo-reject-button>`      | Wraps a button that records "reject all" consent.                                                            |
| `<probo-customize-button>`   | Wraps a button that opens the preference panel.                                                              |
| `<probo-preference-panel>`   | Container for per-category consent toggles.                                                                  |
| `<probo-category-list>`      | Renders a `<template>` once per cookie category. Fills `data-slot="name"` and `data-slot="description"`.     |
| `<probo-category-toggle>`    | Binds the checkbox inside it to the category's consent state.                                                |
| `<probo-cookie-list>`        | Renders a `<template>` once per cookie in the category. Fills `data-slot="name"` and `data-slot="duration"`. |
| `<probo-save-button>`        | Wraps a button that saves the current preference draft.                                                      |
| `<probo-settings-button>`    | Floating button to re-open preferences. Accepts a `position` attribute.                                      |

## Blocking Third-Party Scripts and Elements

Tag any element with `data-cookie-consent="<category>"` to block it until the visitor consents to that category. The SDK will activate matching elements automatically after consent is recorded.

### Scripts

Replace `src` with `data-src` and set `type="text/plain"` to prevent execution:

```html
<script
  type="text/plain"
  data-cookie-consent="analytics"
  data-src="https://example.com/analytics.js"
></script>
```

If the script had a meaningful `type` attribute, preserve it with `data-type`:

```html
<script
  type="text/plain"
  data-type="module"
  data-cookie-consent="analytics"
  data-src="https://example.com/analytics.mjs"
></script>
```

Inline scripts work too:

```html
<script type="text/plain" data-cookie-consent="analytics">
  console.log("This runs only after analytics consent");
</script>
```

### Iframes, Images, and Other Elements

Replace `src` with `data-src` (or `href` with `data-href` for `<link>` tags):

```html
<iframe
  data-cookie-consent="marketing"
  data-src="https://www.youtube.com/embed/dQw4w9WgXcQ"
  width="560"
  height="315"
></iframe>

<img data-cookie-consent="analytics" data-src="https://example.com/pixel.gif" />

<link data-cookie-consent="analytics" data-href="https://example.com/tracker.css" rel="stylesheet" />
```

Supported tags: `<script>`, `<iframe>`, `<img>`, `<video>`, `<audio>`, `<embed>`, `<object>`, `<link>`.

Elements added dynamically after consent is recorded are also activated automatically via a `MutationObserver`.

## License

MIT
