import { useState } from "react";
import { useConfig } from "../hooks/useConfig";
import { configLogger } from "../lib/logger";

export function ConfigTab() {
  const [config, setConfig] = useConfig();
  const [bannerId, setBannerId] = useState(config.bannerId);
  const [baseUrl, setBaseUrl] = useState(config.baseUrl);
  const [gcmEnabled, setGcmEnabled] = useState(config.gcmEnabled);

  const dirty =
    bannerId !== config.bannerId ||
    baseUrl !== config.baseUrl ||
    gcmEnabled !== config.gcmEnabled;

  const save = () => {
    configLogger.debug("[config] save", { bannerId, baseUrl, gcmEnabled });
    setConfig({ bannerId, baseUrl, gcmEnabled });
  };

  return (
    <div>
      <h2>Configuration</h2>
      <p style={{ color: "#666", marginBottom: 16 }}>
        Set the banner ID, base URL, and Google Consent Mode (GCM) option.
        These values are persisted to localStorage and used by every other tab.
      </p>

      <div style={{ marginBottom: 12 }}>
        <label style={{ display: "block", marginBottom: 4, fontWeight: "bold" }}>
          Banner ID
        </label>
        <input
          type="text"
          value={bannerId}
          onChange={(e) => setBannerId(e.target.value)}
          placeholder="e.g. cm9xkz5ab000208jx1yy99abc"
          style={{
            width: "100%",
            maxWidth: 500,
            padding: "6px 8px",
            fontFamily: "monospace",
            fontSize: 14,
            border: "1px solid #ccc",
            boxSizing: "border-box",
          }}
        />
      </div>

      <div style={{ marginBottom: 16 }}>
        <label style={{ display: "block", marginBottom: 4, fontWeight: "bold" }}>
          Base URL
        </label>
        <input
          type="text"
          value={baseUrl}
          onChange={(e) => setBaseUrl(e.target.value)}
          placeholder="e.g. https://cookie-banner.getprobo.com/v1/banners/"
          style={{
            width: "100%",
            maxWidth: 500,
            padding: "6px 8px",
            fontFamily: "monospace",
            fontSize: 14,
            border: "1px solid #ccc",
            boxSizing: "border-box",
          }}
        />
      </div>

      <div style={{ marginBottom: 16 }}>
        <label style={{ display: "flex", alignItems: "center", gap: 8, fontWeight: "bold" }}>
          <input
            type="checkbox"
            checked={gcmEnabled}
            onChange={(e) => setGcmEnabled(e.target.checked)}
          />
          Google Consent Mode (gcm-enabled)
        </label>
        <p style={{ color: "#666", margin: "4px 0 0 24px", fontSize: 13, maxWidth: 520 }}>
          When enabled (default), the SDK sends deny-all{" "}
          <code>consent default</code> on load and maps category choices to{" "}
          <code>gtag</code>/<code>dataLayer</code>. Uncheck if the page already
          manages Consent Mode itself.
        </p>
      </div>

      <button
        onClick={save}
        disabled={!dirty || !bannerId || !baseUrl}
        style={{
          padding: "8px 16px",
          fontWeight: "bold",
          cursor: dirty && bannerId && baseUrl ? "pointer" : "not-allowed",
          opacity: dirty && bannerId && baseUrl ? 1 : 0.5,
        }}
      >
        Save
      </button>

      {config.bannerId && config.baseUrl && (
        <div style={{ marginTop: 24 }}>
          <h3>Current saved config</h3>
          <pre
            style={{
              background: "#f5f5f5",
              padding: 12,
              border: "1px solid #ddd",
              overflow: "auto",
            }}
          >
            {JSON.stringify(config, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
}
