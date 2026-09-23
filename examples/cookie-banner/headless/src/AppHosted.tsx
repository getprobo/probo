// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { useEffect, useRef } from "react";
import {
  registerHeadlessComponents,
  resolveBannerText,
  resolveLayout,
  type BannerConfig,
} from "@probo/cookie-banner/headless";
import {
  bundledConfig,
  DebugPanel,
  EventLog,
  ExampleShell,
  enableNamedLoggers,
  getExampleLogger,
  useEventLog,
} from "@probo/example-cookie-banner-shared";
import { headlessRootHTML } from "./markup";

const headlessLogger = getExampleLogger("headless");

const headlessActions: Record<string, string> = {
  "PROBO-ACKNOWLEDGE-BUTTON": "Acknowledge",
  "PROBO-ACCEPT-BUTTON": "Accept All",
  "PROBO-REJECT-BUTTON": "Reject All",
  "PROBO-CUSTOMIZE-BUTTON": "Customize",
  "PROBO-SAVE-BUTTON": "Save Preferences",
};

function headlessActionLabel(target: EventTarget | null): string | null {
  if (!(target instanceof Element)) {
    return null;
  }

  const host = target.closest(
    "probo-acknowledge-button, probo-accept-button, probo-reject-button, probo-customize-button, probo-save-button",
  );
  if (!host) {
    return null;
  }

  if (host.tagName === "PROBO-REJECT-BUTTON" && host.closest("probo-privacy-choices")) {
    return "Do Not Sell";
  }

  return headlessActions[host.tagName] ?? null;
}

export function App() {
  const { events, pushEvent } = useEventLog();
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    headlessLogger.debug("[headless] registerHeadlessComponents");
    registerHeadlessComponents();
  }, []);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    container.innerHTML = headlessRootHTML(
      bundledConfig.bannerId,
      bundledConfig.baseUrl,
      bundledConfig.gcmEnabled,
    );

    const onClick = (e: Event) => {
      const label = headlessActionLabel(e.target);
      if (label) {
        headlessLogger.debug("[headless] click", label);
      }
    };
    container.addEventListener("click", onClick);

    const root = container.querySelector("probo-cookie-banner-root");
    if (root) {
      root.addEventListener("probo-ready", (e: Event) => {
        const detail = (e as CustomEvent).detail as {
          config?: BannerConfig;
        };
        const bannerConfig = detail?.config;
        enableNamedLoggers();
        headlessLogger.debug("[headless] probo-ready", detail);
        pushEvent("probo-ready", {
          ...detail,
          layout: bannerConfig ? resolveLayout(bannerConfig) : null,
          bannerText: bannerConfig ? resolveBannerText(bannerConfig) : null,
        });
      });
      root.addEventListener("probo-consent", (e: Event) => {
        headlessLogger.debug("[headless] probo-consent", (e as CustomEvent).detail);
        pushEvent("probo-consent", (e as CustomEvent).detail);
      });
    }

    return () => {
      container.removeEventListener("click", onClick);
      container.innerHTML = "";
    };
  }, [pushEvent]);

  return (
    <ExampleShell
      title="@probo/cookie-banner — headless"
      description={
        <>
          Raw headless elements with no themed styling and no TCF. Uses{" "}
          <code>registerHeadlessComponents()</code> and renders raw headless
          elements.{" "}
          <code>gcm-enabled=&quot;{bundledConfig.gcmEnabled ? "true" : "false"}&quot;</code>{" "}
          is baked in at build time. Borders show element boundaries.
        </>
      }
    >
      <div ref={containerRef} style={{ marginTop: 32 }} />
      <DebugPanel bannerId={bundledConfig.bannerId} gcmEnabled={bundledConfig.gcmEnabled} />
      <EventLog events={events} />
    </ExampleShell>
  );
}
