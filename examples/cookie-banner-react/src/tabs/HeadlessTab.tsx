import { useEffect, useRef } from "react";
import {
  registerHeadlessComponents,
  resolveBannerText,
  resolveLayout,
  type BannerConfig,
} from "@probo/cookie-banner/headless";
import { useConfig } from "../hooks/useConfig";
import type { EventEntry } from "../App";

let registered = false;

interface HeadlessTabProps {
  events: EventEntry[];
  pushEvent: (type: string, detail: unknown) => void;
}

export function HeadlessTab({ events, pushEvent }: HeadlessTabProps) {
  const [config] = useConfig();
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!registered) {
      registerHeadlessComponents();
      registered = true;
    }
  }, []);

  useEffect(() => {
    const container = containerRef.current;
    if (!container || !config.bannerId || !config.baseUrl) return;

    container.innerHTML = "";

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

    const root = container.querySelector("probo-cookie-banner-root");
    if (root) {
      root.addEventListener("probo-ready", (e: Event) => {
        const detail = (e as CustomEvent).detail as {
          config?: BannerConfig;
        };
        const bannerConfig = detail?.config;
        pushEvent("probo-ready", {
          ...detail,
          layout: bannerConfig ? resolveLayout(bannerConfig) : null,
          bannerText: bannerConfig ? resolveBannerText(bannerConfig) : null,
        });
      });
      root.addEventListener("probo-consent", (e: Event) =>
        pushEvent("probo-consent", (e as CustomEvent).detail),
      );
    }

    return () => {
      container.innerHTML = "";
    };
  }, [config.bannerId, config.baseUrl, config.gcmEnabled, pushEvent]);

  if (!config.bannerId || !config.baseUrl) {
    return (
      <div>
        <h2>Headless Components</h2>
        <p style={{ color: "tomato" }}>
          Set banner ID and base URL in the Config tab first.
        </p>
      </div>
    );
  }

  return (
    <div>
      <h2>Headless Components</h2>
      <p style={{ color: "#666", marginBottom: 16 }}>
        Uses <code>registerHeadlessComponents()</code> and renders raw headless
        elements with no themed styling.{" "}
        <code>gcm-enabled=&quot;{config.gcmEnabled ? "true" : "false"}&quot;</code>{" "}
        is taken from the Config tab. Borders show element boundaries.
      </p>

      <div ref={containerRef} />

      <h3>Events ({events.length})</h3>
      {events.length === 0 ? (
        <p style={{ color: "#999" }}>No events yet.</p>
      ) : (
        events.map((ev, i) => (
          <div
            key={i}
            style={{
              marginBottom: 8,
              border: "1px solid #ddd",
              padding: 8,
              background: "#f5f5f5",
            }}
          >
            <div style={{ fontWeight: "bold", marginBottom: 4 }}>
              {ev.type}{" "}
              <span style={{ fontWeight: "normal", color: "#999" }}>
                {ev.time}
              </span>
            </div>
            <pre style={{ margin: 0, overflow: "auto" }}>
              {JSON.stringify(ev.detail, null, 2)}
            </pre>
          </div>
        ))
      )}
    </div>
  );
}
