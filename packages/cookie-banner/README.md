# @probo/cookie-banner

A lightweight, dependency-free cookie consent banner built on Web Components. Bundle it with your app as an ES module, use it headless with full UI control, or drop it in with a single script tag. Works with any framework or plain HTML.

Supports opt-in (GDPR, ePrivacy) and opt-out (CCPA/CPRA) consent modes, per-category cookie controls, third-party resource blocking, Google Consent Mode v2, PostHog integration, and multi-language support out of the box.

## Installation

There are three ways to use the SDK:

### Script Tag (IIFE)

No bundler required — add a `<script>` tag and a settings link in the header or footer:

```html
<script
  src="https://cdn.jsdelivr.net/npm/@probo/cookie-banner/dist/cookie-banner.iife.js"
  data-banner-id="YOUR_BANNER_ID"
  data-base-url="https://your-probo-instance.com/api/cookie-banner/v1/"
  data-position="bottom-left"
></script>

<!-- Required: reopen control (header or footer). Auto-fills by regulation. -->
<probo-settings-link></probo-settings-link>
```

This renders a fully styled consent dialog. Place `<probo-settings-link>` in your header or footer so visitors can reopen preferences. Put your default label as children; under CCPA the SDK always replaces it with “Your Privacy Choices” and the official opt-out icon, and the link opens the Privacy Choices panel (sale/sharing opt-out + SPI rights statement).

### ES Module (Themed Banner)

For bundled applications (React, Vue, Svelte, Next.js, etc.):

```bash
npm install @probo/cookie-banner
```

```js
import { registerThemedBanner } from "@probo/cookie-banner";

registerThemedBanner();
```

```html
<probo-cookie-banner
  banner-id="YOUR_BANNER_ID"
  base-url="https://your-probo-instance.com/api/cookie-banner/v1/"
  position="bottom-left"
></probo-cookie-banner>

<!-- Required in header or footer -->
<probo-settings-link></probo-settings-link>
```

See [Theming](https://www.probo.com/docs/product/cookie-banner/theming) to customize colors, fonts, and styling with CSS custom properties.

### Headless Components

For complete control over the consent UI, use the unstyled Web Component building blocks:

```js
import { registerComponents } from "@probo/cookie-banner/headless";

registerComponents();
```

```html
<probo-cookie-banner-root banner-id="YOUR_BANNER_ID" base-url="BASE_URL">
  <probo-banner>
    <div class="my-banner">
      <p>We use cookies to improve your experience.</p>
      <probo-accept-button><button>Accept all</button></probo-accept-button>
      <probo-reject-button><button>Reject all</button></probo-reject-button>
      <probo-customize-button><button>Customize</button></probo-customize-button>
    </div>
  </probo-banner>

  <probo-preference-panel>
    <div class="my-preferences">
      <probo-category-list>
        <template>
          <div class="category">
            <span data-slot="name"></span>
            <span data-slot="description"></span>
            <probo-category-toggle><input type="checkbox" /></probo-category-toggle>
          </div>
        </template>
      </probo-category-list>
      <probo-save-button><button>Save preferences</button></probo-save-button>
    </div>
  </probo-preference-panel>

  <!-- OPT_OUT / CCPA: visible when state === "privacy_choices" -->
  <probo-privacy-choices>
    <div class="my-privacy-choices">
      <!-- title, sale/sharing description, SPI statement -->
      <probo-reject-button><button>Do Not Sell or Share My Personal Information</button></probo-reject-button>
    </div>
  </probo-privacy-choices>
</probo-cookie-banner-root>

<!-- Required in header or footer (outside the root is fine) -->
<probo-settings-link></probo-settings-link>
```

If `<probo-settings-link>` is missing, the SDK emits a `probo-validation` warning (same soft checks used for other headless composition requirements).

Put your default label as children — it is shown for non-CCPA visitors. Under CCPA the SDK always replaces the content with the statutory “Your Privacy Choices” label and official opt-out icon, and opens the Privacy Choices panel (`privacy_choices` state / `<probo-privacy-choices>`).

Style the **host** for font size and color (not inner children). Host styles still apply after CCPA replace; the icon keeps its statutory blue/white fills and scales with `1em`:

```html
<style>
  probo-settings-link {
    font-size: 14px;
    color: #334155;
    text-decoration: underline;
  }
</style>
<probo-settings-link>Cookie settings</probo-settings-link>
```

## Key Features

- **Multi-regulation compliance** — Supports opt-in (GDPR, ePrivacy) and opt-out (CCPA/CPRA) consent modes, Global Privacy Control (GPC) detection, and per-category cookie controls. Under CCPA the banner starts closed; the settings link (“Your Privacy Choices” + official icon) opens the Privacy Choices panel with sale/sharing opt-out and an SPI rights statement.
- **Consent audit trail** — Every consent action is recorded server-side with anonymized IP, user agent, per-category choices, and a reference to the exact banner version the visitor saw.
- **Third-party blocking** — Automatically prevents scripts, iframes, images, and other resources from loading until the visitor grants consent for the matching category.
- **Built-in integrations** — Syncs consent state with Google Consent Mode v2 and PostHog automatically.
- **Multi-language support** — Built-in translations for English, French, German, and Spanish. The SDK auto-detects the visitor's language from the page or browser. The CCPA “Your Privacy Choices” link title stays in English (statutory label).
- **Theming** — Match your brand with CSS custom properties for colors, fonts, border radius, and more. Supports dark mode.

## Documentation

Full documentation is available at **https://www.probo.com/docs/product/cookie-banner/overview**

## License

MIT
