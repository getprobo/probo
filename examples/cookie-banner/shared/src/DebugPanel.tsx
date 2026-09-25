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

import { useEffect, useState, type ReactNode } from "react";
import { getConsent } from "@probo/cookie-banner/consent";
import type { ConsentData } from "@probo/cookie-banner/consent";
import { readGCMSnapshot, type GCMSnapshot } from "./gcm";
import { getExampleLogger } from "./logger";
import { TCFDebugCard } from "./TCFDebugCard";

const debugLogger = getExampleLogger("debug");

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

interface DebugPanelProps {
  bannerId: string;
  gcmEnabled: boolean;
  children?: ReactNode;
}

export function DebugPanel({ bannerId, gcmEnabled, children }: DebugPanelProps) {
  const [snapshot, setSnapshot] = useState<ConsentSnapshot>(readSnapshot);
  const [visitorId, setVisitorId] = useState<string | null>(() =>
    readVisitorId(bannerId),
  );
  const [cookie, setCookie] = useState<string | null>(readCookie);
  const [gcm, setGCM] = useState<GCMSnapshot | null>(readGCMSnapshot);

  useEffect(() => {
    const mgr = getConsent();
    return mgr.subscribe(() => {
      const next = readSnapshot();
      debugLogger.debug("[debug] consent", next);
      setSnapshot(next);
      setVisitorId(readVisitorId(bannerId));
      setCookie(readCookie());
      setGCM(readGCMSnapshot());
    });
  }, [bannerId]);

  useEffect(() => {
    setVisitorId(readVisitorId(bannerId));
    setCookie(readCookie());
    setGCM(readGCMSnapshot());
  }, [bannerId, gcmEnabled]);

  useEffect(() => {
    const id = window.setInterval(() => setGCM(readGCMSnapshot()), 1000);
    return () => window.clearInterval(id);
  }, []);

  return (
    <section style={{ marginTop: 32 }}>
      <h2>State</h2>

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
            (localStorage: probo_consent:{bannerId || "?"}:vid)
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

      <div
        style={{
          border: "1px solid #ccc",
          padding: 12,
          background: "#fafafa",
          marginBottom: 16,
        }}
      >
        <h3 style={{ marginTop: 0 }}>Google Consent Mode</h3>
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
            {
              enabled: gcmEnabled,
              command: gcm?.command ?? null,
            },
            null,
            2,
          )}
        </pre>

        {!gcm || Object.keys(gcm.signals).length === 0 ? (
          <p style={{ color: "#999", margin: 0 }}>
            No dataLayer consent calls yet.
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
                  Signal
                </th>
                <th
                  style={{
                    textAlign: "left",
                    padding: "4px 0",
                    borderBottom: "1px solid #ccc",
                  }}
                >
                  state
                </th>
              </tr>
            </thead>
            <tbody>
              {Object.entries(gcm.signals).map(([signal, state]) => (
                <tr key={signal}>
                  <td style={{ padding: "4px 16px 4px 0" }}>{signal}</td>
                  <td
                    style={{
                      padding: "4px 0",
                      color: state === "granted" ? "green" : state === "denied" ? "red" : undefined,
                      fontWeight: "bold",
                    }}
                  >
                    {state}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <TCFDebugCard />
      {children}
    </section>
  );
}
