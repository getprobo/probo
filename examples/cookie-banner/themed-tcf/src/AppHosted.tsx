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
import { startTCF } from "@probo/cookie-banner-tcf";
import {
  bundledConfig,
  DebugPanel,
  EventLog,
  ExampleShell,
  enableNamedLoggers,
  getExampleLogger,
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

const themedLogger = getExampleLogger("themed-tcf");

let started = false;

export function App() {
  const { events, pushEvent } = useEventLog();
  const posthogStatus = usePosthogStatus();
  const [manualPing, setManualPing] = useState<string | null>(null);

  useEffect(() => {
    if (!started) {
      themedLogger.debug("[themed-tcf] startTCF, registerCookieBanner");
      startTCF();
      registerCookieBanner();
      started = true;
    }
    initPosthog();
  }, []);

  const attachListeners = useCallback(
    (el: HTMLElement | null) => {
      if (!el) return;

      el.addEventListener("probo-ready", (e: Event) => {
        const detail = (e as CustomEvent).detail as {
          config?: BannerConfig;
        };
        if (detail?.config) {
          configurePosthogFromBanner(detail.config);
        }
        enableNamedLoggers();
        themedLogger.debug("[themed-tcf] probo-ready", (e as CustomEvent).detail);
        pushEvent("probo-ready", (e as CustomEvent).detail);
      });
      el.addEventListener("probo-consent", (e: Event) => {
        themedLogger.debug("[themed-tcf] probo-consent", (e as CustomEvent).detail);
        pushEvent("probo-consent", (e as CustomEvent).detail);
      });
    },
    [pushEvent],
  );

  const sendPing = useCallback(() => {
    if (posthog.has_opted_out_capturing()) return;
    themedLogger.debug("[themed-tcf] posthog ping");
    posthog.capture("themed_tcf_manual_ping", { source: "example" });
    setManualPing(new Date().toISOString());
  }, []);

  return (
    <ExampleShell
      title="@probo/cookie-banner — themed TCF"
      description={
        <>
          IAB TCF first-layer path. Use this app with the IAB CMP validator.
          Uses <code>installTCFStub()</code>, <code>startTCF()</code>, and{" "}
          <code>registerCookieBanner()</code>, then renders{" "}
          <code>&lt;probo-cookie-banner&gt;</code> with{" "}
          <code>gcm-enabled=&quot;{bundledConfig.gcmEnabled ? "true" : "false"}&quot;</code>.
          TCF disclosures replace the category banner when the configured banner
          has the capability on.
        </>
      }
    >
      <probo-cookie-banner
        key={`${bundledConfig.bannerId}:${bundledConfig.baseUrl}:${bundledConfig.gcmEnabled}`}
        ref={attachListeners}
        banner-id={bundledConfig.bannerId}
        base-url={bundledConfig.baseUrl}
        position="bottom-right"
        gcm-enabled={bundledConfig.gcmEnabled ? "true" : "false"}
      />

      <DebugPanel bannerId={bundledConfig.bannerId} gcmEnabled={bundledConfig.gcmEnabled}>
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
