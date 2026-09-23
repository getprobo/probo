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
  ConfigForm,
  DebugPanel,
  EventLog,
  ExampleShell,
  enableNamedLoggers,
  getExampleLogger,
  useConfig,
  useEventLog,
} from "@probo/example-cookie-banner-shared";

const headlessLogger = getExampleLogger("headless");

const headlessActions: Record<string, string> = {
  "PROBO-ACKNOWLEDGE-BUTTON": "Acknowledge",
  "PROBO-ACCEPT-BUTTON": "Accept All",
  "PROBO-REJECT-BUTTON": "Reject All",
  "PROBO-CUSTOMIZE-BUTTON": "Customize",
  "PROBO-SAVE-BUTTON": "Save Preferences",
  "PROBO-SETTINGS-LINK": "Cookie settings",
};

function headlessActionLabel(target: EventTarget | null): string | null {
  if (!(target instanceof Element)) {
    return null;
  }

  const host = target.closest(
    "probo-acknowledge-button, probo-accept-button, probo-reject-button, probo-customize-button, probo-save-button, probo-settings-link",
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
  const [config] = useConfig();
  const { events, pushEvent } = useEventLog();
  const containerRef = useRef<HTMLDivElement>(null);
  const ready = Boolean(config.bannerId && config.baseUrl);

  useEffect(() => {
    headlessLogger.debug("[headless] registerHeadlessComponents");
    registerHeadlessComponents();
  }, []);

  useEffect(() => {
    const container = containerRef.current;
    if (!container || !config.bannerId || !config.baseUrl) return;

    container.innerHTML = `
      <style>probo-banner, probo-preference-panel, probo-privacy-choices { display: block !important; }</style>
      <probo-cookie-banner-root banner-id="${config.bannerId}" base-url="${config.baseUrl}" gcm-enabled="${config.gcmEnabled ? "true" : "false"}">
        <probo-banner>
          <div style="border:2px solid #333;padding:12px;margin-bottom:8px;">
            <strong>[probo-banner]</strong>
            <div style="margin-top:8px;">
              <probo-acknowledge-button><button>Acknowledge</button></probo-acknowledge-button>
              <probo-accept-button><button style="margin-left:8px;">Accept All</button></probo-accept-button>
              <probo-reject-button><button style="margin-left:8px;">Reject All</button></probo-reject-button>
              <probo-customize-button><button style="margin-left:8px;">Customize</button></probo-customize-button>
            </div>
          </div>
        </probo-banner>

        <probo-preference-panel>
          <div style="border:2px dashed #666;padding:12px;margin-bottom:8px;">
            <strong>[probo-preference-panel]</strong>
            <probo-category-list>
              <template>
                <div style="border:1px solid #aaa;padding:8px;margin:4px 0;">
                  <span data-slot="name" style="font-weight:bold;"></span>:
                  <span data-slot="description"></span>
                  <probo-category-toggle>
                    <label style="margin-left:8px;"><input type="checkbox" /> toggle</label>
                  </probo-category-toggle>
                  <probo-cookie-list hidden>
                    <template>
                      <div style="padding:4px 0 4px 16px;font-size:13px;">
                        <span data-slot="name" style="font-weight:bold;"></span>
                        &mdash; <span data-slot="description"></span>
                      </div>
                    </template>
                  </probo-cookie-list>
                </div>
              </template>
            </probo-category-list>
            <div style="margin-top:8px;">
              <probo-accept-button><button>Accept All</button></probo-accept-button>
              <probo-reject-button><button style="margin-left:8px;">Reject All</button></probo-reject-button>
              <probo-save-button><button style="margin-left:8px;">Save Preferences</button></probo-save-button>
            </div>
          </div>
        </probo-preference-panel>

        <probo-privacy-choices>
          <div style="border:2px solid #1d4ed8;padding:12px;margin-bottom:8px;">
            <strong>[probo-privacy-choices]</strong>
            <p style="margin:8px 0;font-size:14px;">
              Right to opt out of sale/sharing and right to limit sensitive
              personal information (CCPA).
            </p>
            <probo-reject-button>
              <button>Do Not Sell or Share My Personal Information</button>
            </probo-reject-button>
          </div>
        </probo-privacy-choices>
      </probo-cookie-banner-root>
      <p style="margin-top:12px;">
        <probo-settings-link>Cookie settings</probo-settings-link>
      </p>
    `;

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
  }, [config.bannerId, config.baseUrl, config.gcmEnabled, pushEvent]);

  return (
    <ExampleShell
      current="headless"
      title="@probo/cookie-banner — headless"
      description="Raw headless elements with no themed styling and no TCF."
    >
      <ConfigForm />

      <section style={{ marginTop: 32 }}>
        <h2>Headless Components</h2>
        {ready ? (
          <>
            <p style={{ color: "#666", marginBottom: 16 }}>
              Uses <code>registerHeadlessComponents()</code> and renders raw
              headless elements.{" "}
              <code>gcm-enabled=&quot;{config.gcmEnabled ? "true" : "false"}&quot;</code>{" "}
              comes from the configuration section. Borders show element
              boundaries.
            </p>
            <div ref={containerRef} />
            <EventLog events={events} />
          </>
        ) : (
          <p style={{ color: "tomato" }}>
            Set banner ID and base URL in the configuration section first.
          </p>
        )}
      </section>

      <div style={{ marginTop: 32 }}>
        <DebugPanel />
      </div>
    </ExampleShell>
  );
}
