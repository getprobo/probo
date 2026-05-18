import { useEffect, useState } from "react";
import { getConsent } from "@probo/cookie-banner/consent";
import type { ConsentData } from "@probo/cookie-banner/consent";
import { useConfig } from "../hooks/useConfig";

interface ConsentSnapshot {
  ready: boolean;
  hasResponse: boolean;
  data: ConsentData;
}

function readSnapshot(): ConsentSnapshot {
  const mgr = getConsent();
  return {
    ready: mgr.ready,
    hasResponse: mgr.hasResponse,
    data: mgr.getAll(),
  };
}

function readVisitorId(bannerId: string): string | null {
  if (!bannerId) return null;
  try {
    return localStorage.getItem(`probo_consent:${bannerId}:vid`);
  } catch {
    return null;
  }
}

function readCookie(): string | null {
  try {
    const prefix = "probo_consent=";
    const entry = document.cookie
      .split("; ")
      .find((c) => c.startsWith(prefix));
    if (!entry) return null;
    return decodeURIComponent(entry.substring(prefix.length));
  } catch {
    return null;
  }
}

export function DebugTab() {
  const [config] = useConfig();
  const [snapshot, setSnapshot] = useState<ConsentSnapshot>(readSnapshot);
  const [visitorId, setVisitorId] = useState<string | null>(() =>
    readVisitorId(config.bannerId),
  );
  const [cookie, setCookie] = useState<string | null>(readCookie);

  useEffect(() => {
    const mgr = getConsent();
    return mgr.subscribe(() => {
      setSnapshot(readSnapshot());
      setVisitorId(readVisitorId(config.bannerId));
      setCookie(readCookie());
    });
  }, [config.bannerId]);

  useEffect(() => {
    setVisitorId(readVisitorId(config.bannerId));
    setCookie(readCookie());
  }, [config.bannerId]);

  return (
    <div>
      <h2>Debug</h2>

      <div
        style={{
          border: "1px solid #ccc",
          padding: 12,
          background: "#fafafa",
          marginBottom: 16,
        }}
      >
        <h3 style={{ marginTop: 0 }}>getConsent() State</h3>
        <pre
          style={{
            background: "#f0f0f0",
            padding: 8,
            border: "1px solid #ddd",
            overflow: "auto",
            margin: "0 0 12px 0",
          }}
        >
          {JSON.stringify(
            { ready: snapshot.ready, hasResponse: snapshot.hasResponse },
            null,
            2,
          )}
        </pre>

        {Object.keys(snapshot.data).length === 0 ? (
          <p style={{ color: "#999", margin: 0 }}>
            No consent data yet. Interact with the banner to generate consent.
          </p>
        ) : (
          <table
            style={{
              borderCollapse: "collapse",
              fontFamily: "monospace",
              fontSize: 14,
            }}
          >
            <thead>
              <tr>
                <th
                  style={{
                    textAlign: "left",
                    padding: "4px 16px 4px 0",
                    borderBottom: "1px solid #ccc",
                  }}
                >
                  Category
                </th>
                <th
                  style={{
                    textAlign: "left",
                    padding: "4px 0",
                    borderBottom: "1px solid #ccc",
                  }}
                >
                  has()
                </th>
              </tr>
            </thead>
            <tbody>
              {Object.entries(snapshot.data).map(([cat, granted]) => (
                <tr key={cat}>
                  <td style={{ padding: "4px 16px 4px 0" }}>{cat}</td>
                  <td
                    style={{
                      padding: "4px 0",
                      color: granted ? "green" : "red",
                      fontWeight: "bold",
                    }}
                  >
                    {String(granted)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div
        style={{
          border: "1px solid #ccc",
          padding: 12,
          background: "#fafafa",
          marginBottom: 16,
        }}
      >
        <h3 style={{ marginTop: 0 }}>Storage</h3>

        <div style={{ marginBottom: 12 }}>
          <strong>Visitor ID</strong>{" "}
          <span style={{ fontFamily: "monospace", fontSize: 13, color: "#666" }}>
            (localStorage: probo_consent:{config.bannerId || "?"}:vid)
          </span>
          <pre
            style={{
              background: "#f0f0f0",
              padding: 8,
              border: "1px solid #ddd",
              overflow: "auto",
              margin: "4px 0 0 0",
            }}
          >
            {visitorId ?? "(not set)"}
          </pre>
        </div>

        <div>
          <strong>probo_consent cookie</strong>
          <pre
            style={{
              background: "#f0f0f0",
              padding: 8,
              border: "1px solid #ddd",
              overflow: "auto",
              margin: "4px 0 0 0",
            }}
          >
            {cookie
              ? JSON.stringify(JSON.parse(cookie), null, 2)
              : "(not set)"}
          </pre>
        </div>
      </div>
    </div>
  );
}
