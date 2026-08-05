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

import type { PosthogStatus } from "../../lib/posthog";

interface PosthogPanelProps {
  status: PosthogStatus;
  manualPing: string | null;
  onSendPing: () => void;
}

export function PosthogPanel({ status, manualPing, onSendPing }: PosthogPanelProps) {
  return (
    <div
      style={{
        border: "1px solid #ccc",
        padding: 12,
        background: "#fafafa",
        marginBottom: 16,
      }}
    >
      <h3 style={{ marginTop: 0 }}>PostHog</h3>
      <p style={{ color: "#666", margin: "0 0 8px 0", fontSize: 13 }}>
        Init uses <code>cookieless_mode: &quot;on_reject&quot;</code> and sets{" "}
        <code>opt_out_capturing_by_default</code> from the analytics consent
        snapshot. Accept analytics → <code>opt_in_capturing()</code> +{" "}
        <code>identify()</code>; feature flags stay forced off until both
        happen. Reject → <code>opt_out_capturing()</code> +{" "}
        <code>reset()</code> (flag treated as false).
      </p>
      <table
        style={{
          borderCollapse: "collapse",
          fontFamily: "monospace",
          fontSize: 13,
          marginBottom: 8,
        }}
      >
        <tbody>
          <Row label="initialized" value={String(status.initialized)} ok={status.initialized} />
          <Row label="consent mode" value={status.consentMode ?? "(pending)"} />
          <Row
            label="consent status"
            value={status.consentStatus ?? "(pending)"}
            ok={status.consentStatus === "granted"}
            warn={status.consentStatus === "denied"}
          />
          <Row
            label="opted in"
            value={String(status.optedIn)}
            ok={status.optedIn}
          />
          <Row
            label="opted out"
            value={String(status.optedOut)}
            warn={status.optedOut}
          />
          <Row
            label="identified"
            value={String(status.identified)}
            ok={status.identified}
          />
          <Row label="distinct_id" value={status.distinctId ?? "(none)"} />
          <Row label="flag key" value={status.featureFlagKey} />
          <Row
            label="flag enabled"
            value={String(status.featureFlagEnabled)}
            ok={status.featureFlagEnabled}
          />
        </tbody>
      </table>
      <button
        onClick={onSendPing}
        disabled={!status.initialized || status.optedOut}
        style={{ padding: "6px 12px", fontSize: 13 }}
      >
        Capture test event
      </button>
      {manualPing && (
        <span style={{ marginLeft: 12, color: "#666", fontSize: 13 }}>
          last sent: {manualPing}
        </span>
      )}
    </div>
  );
}

interface RowProps {
  label: string;
  value: string;
  ok?: boolean;
  warn?: boolean;
}

function Row({ label, value, ok, warn }: RowProps) {
  const color = ok ? "green" : warn ? "tomato" : undefined;
  return (
    <tr>
      <td style={{ padding: "2px 16px 2px 0", color: "#666" }}>{label}</td>
      <td style={{ padding: "2px 0", color, fontWeight: ok || warn ? "bold" : "normal" }}>
        {value}
      </td>
    </tr>
  );
}
