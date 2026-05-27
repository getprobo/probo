import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from "react";
import posthog from "posthog-js";
import { registerCookieBanner, type BannerConfig } from "@probo/cookie-banner";
import { useConfig } from "../hooks/useConfig";
import {
  configurePosthogFromBanner,
  getPosthogStatus,
  initPosthog,
  subscribePosthogStatus,
  type PosthogStatus,
} from "../lib/posthog";
import { PosthogPanel } from "./_components/PosthogPanel";
import type { EventEntry } from "../App";

let registered = false;

interface ThemedBannerTabProps {
  events: EventEntry[];
  pushEvent: (type: string, detail: unknown) => void;
}

export function ThemedBannerTab({ events, pushEvent }: ThemedBannerTabProps) {
  const [config] = useConfig();
  const elRef = useRef<HTMLElement | null>(null);
  const posthogStatus = usePosthogStatus();
  const [manualPing, setManualPing] = useState<string | null>(null);

  useEffect(() => {
    if (!registered) {
      registerCookieBanner();
      registered = true;
    }
    initPosthog();
  }, []);

  const attachListeners = useCallback(
    (el: HTMLElement | null) => {
      elRef.current = el;
      if (!el) return;

      el.addEventListener("probo-ready", (e: Event) => {
        const detail = (e as CustomEvent).detail as {
          config?: BannerConfig;
        };
        if (detail?.config) {
          configurePosthogFromBanner(detail.config);
        }
        pushEvent("probo-ready", (e as CustomEvent).detail);
      });
      el.addEventListener("probo-consent", (e: Event) =>
        pushEvent("probo-consent", (e as CustomEvent).detail),
      );
    },
    [pushEvent],
  );

  const sendPing = useCallback(() => {
    // Defensive: `PosthogPanel` already disables the button while opted out,
    // and pure telemetry pings like this one are fine to ship cookielessly
    // (PostHog's `cookieless_mode` is designed exactly for that). Only gate
    // manual captures when the event carries user-specific data — e.g. an
    // identified `distinct_id`, an email, a workspace name — that must not
    // leave the browser without consent. We keep the guard here purely to
    // make the example fail closed if the panel is reused without its
    // disabled-state wiring.
    if (posthog.has_opted_out_capturing()) return;
    posthog.capture("themed_tab_manual_ping", { source: "example" });
    setManualPing(new Date().toISOString());
  }, []);

  if (!config.bannerId || !config.baseUrl) {
    return (
      <div>
        <h2>Themed Banner</h2>
        <p style={{ color: "tomato" }}>
          Set banner ID and base URL in the Config tab first.
        </p>
      </div>
    );
  }

  return (
    <div>
      <h2>Themed Banner</h2>
      <p style={{ color: "#666", marginBottom: 16 }}>
        Uses <code>registerCookieBanner()</code> and renders{" "}
        <code>&lt;probo-cookie-banner&gt;</code>. The banner appears in the
        bottom-right corner.
      </p>

      <PosthogPanel
        status={posthogStatus}
        manualPing={manualPing}
        onSendPing={sendPing}
      />

      <probo-cookie-banner
        ref={attachListeners}
        banner-id={config.bannerId}
        base-url={config.baseUrl}
        position="bottom-right"
      />

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

function usePosthogStatus(): PosthogStatus {
  return useSyncExternalStore(subscribePosthogStatus, getPosthogStatus, getPosthogStatus);
}
