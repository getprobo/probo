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
        Init is deferred until the banner config arrives, then{" "}
        <code>cookieless_mode</code> and{" "}
        <code>opt_out_capturing_by_default</code> are derived from the consent
        snapshot for the category flagged with <code>posthog_consent</code>.
        The snapshot already encodes both the regulation's default
        (<code>OPT_IN</code> → off, <code>OPT_OUT</code> → on) and any
        persisted answer from a prior visit, so a returning visitor who
        accepted boots straight into <code>"on_reject"</code> +{" "}
        <code>false</code> while one who rejected (or a fresh{" "}
        <code>OPT_IN</code> visitor) boots into <code>"always"</code> +{" "}
        <code>true</code>. The <code>getConsent().subscribe()</code> callback
        then flips <code>opt_in_capturing()</code> /{" "}
        <code>opt_out_capturing()</code> on subsequent banner actions.
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
            label="opted in"
            value={String(status.optedIn)}
            ok={status.optedIn}
          />
          <Row
            label="opted out"
            value={String(status.optedOut)}
            warn={status.optedOut}
          />
          <Row label="distinct_id" value={status.distinctId ?? "(none)"} />
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
