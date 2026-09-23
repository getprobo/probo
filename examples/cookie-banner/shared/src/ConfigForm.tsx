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

import { useState } from "react";
import { getExampleLogger } from "./logger";
import { useConfig } from "./useConfig";

const configLogger = getExampleLogger("config");

export function ConfigForm() {
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
    <section>
      <h2>Configuration</h2>
      <p style={{ color: "#666", marginBottom: 16 }}>
        Banner ID, base URL, and Google Consent Mode. Values persist in
        localStorage and are shared by the themed, themed TCF, and headless
        examples.
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
    </section>
  );
}
