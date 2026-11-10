# Cookie banner examples

Three Vite apps that each run a single banner instance. Shared configuration,
debug, and event UI live in `@probo/example-cookie-banner-shared`.

| App | Workspace | URL |
| --- | --- | --- |
| Themed | `@probo/example-cookie-banner-themed` | http://localhost:5180 |
| Themed TCF | `@probo/example-cookie-banner-themed-tcf` | http://localhost:5181 |
| Headless | `@probo/example-cookie-banner-headless` | http://localhost:5182 |

The themed TCF app is the IAB CMP validator path: it installs the `__tcfapi`
stub before React boots, then calls `startTCF()` and `registerCookieBanner()`.
The themed app never loads `@probo/cookie-banner-tcf`. The headless app only
registers headless components.

Banner ID, base URL, and GCM persist in each app's `localStorage` under
`probo-example-config` when the playground is configurable. The apps run
on different ports, so each keeps its own copy.

When both `PUBLIC_COOKIE_BANNER_ID` and
`PUBLIC_COOKIE_BANNER_API_BASE_URL` are non-empty at build time, Vite
inlines them and the app compiles without the configuration form or
`localStorage` config. Leave the ID blank for the local playground.

The IAB-facing static app is the themed TCF build:

```bash
npm -w @probo/example-cookie-banner-themed-tcf run build
```

## GitHub Pages

`.github/workflows/pages-cookie-banner-themed-tcf.yaml` publishes that
build to this repository's GitHub Pages site. Set Pages source to
GitHub Actions, then add these variables on the
`cookie-banner-example-pages` environment:

- `PUBLIC_COOKIE_BANNER_ID`
- `PUBLIC_COOKIE_BANNER_API_BASE_URL`

Both must be non-empty so Vite bakes the hosted banner in. An empty
id compiles the playground form, which this workflow rejects.

## Two sites on one domain

macOS resolves `*.localhost` to `127.0.0.1`. `shop.example.localhost` and
`blog.example.localhost` share the CMP domain `example.localhost`.
`shop.localhost` and `blog.localhost` do not.

`PUBLIC_EXAMPLE_DOMAIN` defaults to `example.localhost`. The themed app
is the reader at `http://shop.<domain>:5180`. The themed TCF app is the
writer at `http://blog.<domain>:5181` and sets
`probo_example=1; Domain=<domain>; Path=/`. The headless app stays on
`http://localhost:5182`.

A console scan of these hosts fails: the scanner refuses loopback. This
exercises the visitor SDK. Open the writer first, then the reader, in
the same browser. The reader's tracker source becomes Script with a link
back to the writer.

1. Create both banners in one organization. The origins must be exactly
   `http://shop.example.localhost:5180` and
   `http://blog.example.localhost:5181`. The API only reflects an origin
   that matches the banner.
2. Put the ids in `PUBLIC_THEMED_BANNER_ID` and
   `PUBLIC_THEMED_TCF_BANNER_ID`, or paste them into each playground.
   An empty id keeps the form. `PUBLIC_COOKIE_BANNER_ID` is still the
   fallback for a single-banner playground.

## Run

1. Copy `.env.example` to `.env` in this directory and fill in values.
2. From the repo root:

```bash
npm -w @probo/example-cookie-banner-themed run dev
npm -w @probo/example-cookie-banner-themed-tcf run dev
npm -w @probo/example-cookie-banner-headless run dev
```
