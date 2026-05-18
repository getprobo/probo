import { useCallback, useEffect, useRef } from "react";
import { registerCookieBanner } from "@probo/cookie-banner";
import { useConfig } from "../hooks/useConfig";
import type { EventEntry } from "../App";

let registered = false;

interface ThemedBannerTabProps {
  events: EventEntry[];
  pushEvent: (type: string, detail: unknown) => void;
}

export function ThemedBannerTab({ events, pushEvent }: ThemedBannerTabProps) {
  const [config] = useConfig();
  const elRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!registered) {
      registerCookieBanner();
      registered = true;
    }
  }, []);

  const attachListeners = useCallback(
    (el: HTMLElement | null) => {
      elRef.current = el;
      if (!el) return;

      el.addEventListener("probo-ready", (e: Event) =>
        pushEvent("probo-ready", (e as CustomEvent).detail),
      );
      el.addEventListener("probo-consent", (e: Event) =>
        pushEvent("probo-consent", (e as CustomEvent).detail),
      );
    },
    [pushEvent],
  );

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
