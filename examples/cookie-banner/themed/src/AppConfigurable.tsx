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

import { useCallback, useEffect, useState, useSyncExternalStore } from "react";
import posthog from "posthog-js";
import { registerCookieBanner, type BannerConfig } from "@probo/cookie-banner";
import {
  ConfigForm,
  DebugPanel,
  EventLog,
  ExampleShell,
  enableNamedLoggers,
  getExampleLogger,
  isValidBannerApiBaseUrl,
  useConfig,
  useEventLog,
} from "@probo/example-cookie-banner-shared";
import {
  configurePosthogFromBanner,
  getPosthogStatus,
  initPosthog,
  PosthogPanel,
  subscribePosthogStatus,
  type PosthogStatus,
} from "@probo/example-cookie-banner-shared/posthog";

const themedLogger = getExampleLogger("themed");

export function App() {
  const [config] = useConfig();
  const { events, pushEvent } = useEventLog();
  const posthogStatus = usePosthogStatus();
  const [manualPing, setManualPing] = useState<string | null>(null);

  useEffect(() => {
    themedLogger.debug("[themed] registerCookieBanner");
    registerCookieBanner();
    initPosthog();
  }, []);

  const attachListeners = useCallback(
    (el: HTMLElement | null) => {
      if (!el) return undefined;

      const onReady = (e: Event): void => {
        const detail = (e as CustomEvent).detail as {
          config?: BannerConfig;
        };
        if (detail?.config) {
          configurePosthogFromBanner(detail.config);
        }
        enableNamedLoggers();
        themedLogger.debug("[themed] probo-ready", (e as CustomEvent).detail);
        pushEvent("probo-ready", (e as CustomEvent).detail);
      };
      const onConsent = (e: Event): void => {
        themedLogger.debug("[themed] probo-consent", (e as CustomEvent).detail);
        pushEvent("probo-consent", (e as CustomEvent).detail);
      };
      el.addEventListener("probo-ready", onReady);
      el.addEventListener("probo-consent", onConsent);
      return () => {
        el.removeEventListener("probo-ready", onReady);
        el.removeEventListener("probo-consent", onConsent);
      };
    },
    [pushEvent],
  );

  const sendPing = useCallback(() => {
    if (posthog.has_opted_out_capturing()) return;
    themedLogger.debug("[themed] posthog ping");
    posthog.capture("themed_tab_manual_ping", { source: "example" });
    setManualPing(new Date().toISOString());
  }, []);

  const ready = Boolean(config.bannerId && isValidBannerApiBaseUrl(config.baseUrl));

  return (
    <ExampleShell
      title="@probo/cookie-banner — themed"
      description={
        <>
          Category banner only. This app never installs the TCF stub or starts
          TCF. Calls <code>registerCookieBanner()</code> and renders{" "}
          <code>&lt;probo-cookie-banner&gt;</code> with{" "}
          <code>gcm-enabled=&quot;{config.gcmEnabled ? "true" : "false"}&quot;</code>.
          The banner appears in the bottom-right corner.
        </>
      }
    >
      <ConfigForm />

      {ready ? (
        <probo-cookie-banner
          key={`${config.bannerId}:${config.baseUrl}:${config.gcmEnabled}`}
          ref={attachListeners}
          banner-id={config.bannerId}
          base-url={config.baseUrl}
          position="bottom-right"
          gcm-enabled={config.gcmEnabled ? "true" : "false"}
        />
      ) : (
        <p style={{ color: "tomato", marginTop: 32 }}>
          {config.bannerId && config.baseUrl
            ? "Base URL must be an absolute http(s) URL."
            : "Set banner ID and base URL in the configuration section first."}
        </p>
      )}

      <DebugPanel bannerId={config.bannerId} gcmEnabled={config.gcmEnabled}>
        <PosthogPanel
          status={posthogStatus}
          manualPing={manualPing}
          onSendPing={sendPing}
        />
        {posthogStatus.featureFlagEnabled && (
          <div
            style={{
              border: "2px solid #2563eb",
              padding: 12,
              marginBottom: 16,
              background: "#eff6ff",
            }}
          >
            <h3 style={{ marginTop: 0, marginBottom: 4 }}>Beta panel</h3>
            <p style={{ margin: 0, color: "#334155", fontSize: 14 }}>
              Visible only when PostHog consent is granted, the demo user is
              identified, and feature flag{" "}
              <code>{posthogStatus.featureFlagKey}</code> is on.
            </p>
          </div>
        )}
      </DebugPanel>

      <EventLog events={events} />
    </ExampleShell>
  );
}

function usePosthogStatus(): PosthogStatus {
  return useSyncExternalStore(subscribePosthogStatus, getPosthogStatus, getPosthogStatus);
}
