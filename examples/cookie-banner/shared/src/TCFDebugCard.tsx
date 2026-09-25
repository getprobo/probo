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

import { useEffect, useState } from "react";
import { getConsent } from "@probo/cookie-banner/consent";

type TCFAPI = (
  command: string,
  version: number,
  callback: (data: unknown, success: boolean) => void,
  parameter?: unknown,
) => void;

interface TCFPing {
  gdprApplies?: boolean;
  cmpLoaded?: boolean;
  cmpStatus?: string;
  displayStatus?: string;
  apiVersion?: string;
  cmpId?: number;
  cmpVersion?: number;
  gvlVersion?: number;
  tcfPolicyVersion?: number;
}

interface TCFPurposeMap {
  consents?: Record<string, boolean>;
  legitimateInterests?: Record<string, boolean>;
}

interface TCFData {
  tcString?: string;
  tcfPolicyVersion?: number;
  cmpId?: number;
  cmpVersion?: number;
  gdprApplies?: boolean;
  eventStatus?: string;
  cmpStatus?: string;
  displayStatus?: string;
  isServiceSpecific?: boolean;
  useNonStandardTexts?: boolean;
  publisherCC?: string;
  purpose?: TCFPurposeMap;
  vendor?: TCFPurposeMap;
  specialFeatureOptins?: Record<string, boolean>;
}

interface TCFSnapshot {
  ping: TCFPing | null;
  tcData: TCFData | null;
}

function tcfapi(): TCFAPI | undefined {
  const api = (window as unknown as { __tcfapi?: TCFAPI }).__tcfapi;
  return typeof api === "function" ? api : undefined;
}

function callTCF(command: string): Promise<unknown | null> {
  return new Promise((resolve) => {
    const api = tcfapi();
    if (!api) {
      resolve(null);
      return;
    }

    let settled = false;
    const timer = window.setTimeout(() => {
      if (!settled) {
        settled = true;
        resolve(null);
      }
    }, 250);

    api(command, 2, (data, success) => {
      if (settled) {
        return;
      }
      settled = true;
      window.clearTimeout(timer);
      resolve(success ? data : null);
    });
  });
}

function isLiveCMP(ping: TCFPing | null): boolean {
  return ping?.cmpLoaded === true && ping.displayStatus !== "disabled";
}

async function readTCF(): Promise<TCFSnapshot | null> {
  if (!tcfapi()) {
    return null;
  }

  const pingRaw = await callTCF("ping");
  const ping = pingRaw && typeof pingRaw === "object" ? (pingRaw as TCFPing) : null;
  if (!isLiveCMP(ping)) {
    return { ping, tcData: null };
  }

  const tcDataRaw = await callTCF("getTCData");
  return {
    ping,
    tcData: tcDataRaw && typeof tcDataRaw === "object" ? (tcDataRaw as TCFData) : null,
  };
}

function grantedIds(map?: Record<string, boolean>): number[] {
  if (!map) {
    return [];
  }
  return Object.entries(map)
    .filter(([, granted]) => granted)
    .map(([id]) => Number(id))
    .filter((id) => Number.isFinite(id))
    .sort((a, b) => a - b);
}

export function TCFDebugCard() {
  const [snapshot, setSnapshot] = useState<TCFSnapshot | null>(null);

  useEffect(() => {
    let cancelled = false;

    const refresh = (): void => {
      void readTCF().then((next) => {
        if (!cancelled) {
          setSnapshot(next);
        }
      });
    };

    refresh();
    const unsubscribe = getConsent().subscribe(refresh);
    const id = window.setInterval(refresh, 1000);
    return () => {
      cancelled = true;
      unsubscribe();
      window.clearInterval(id);
    };
  }, []);

  if (!snapshot) {
    return null;
  }

  const { ping, tcData } = snapshot;
  const tcString = tcData?.tcString ?? "";
  const purposeConsents = grantedIds(tcData?.purpose?.consents);
  const purposeLI = grantedIds(tcData?.purpose?.legitimateInterests);
  const vendorConsents = grantedIds(tcData?.vendor?.consents);
  const vendorLI = grantedIds(tcData?.vendor?.legitimateInterests);
  const specialFeatures = grantedIds(tcData?.specialFeatureOptins);

  return (
    <div
      style={{
        border: "1px solid #ccc",
        padding: 12,
        background: "#fafafa",
        marginBottom: 16,
      }}
    >
      <h3 style={{ marginTop: 0 }}>IAB TCF</h3>
      <p style={{ color: "#666", margin: "0 0 8px 0", fontSize: 13 }}>
        Live <code>__tcfapi</code> <code>ping</code> and <code>getTCData</code>.
      </p>
      <table
        style={{
          borderCollapse: "collapse",
          fontFamily: "monospace",
          fontSize: 13,
          marginBottom: 12,
        }}
      >
        <tbody>
          <Row label="cmpLoaded" value={fmt(ping?.cmpLoaded)} ok={ping?.cmpLoaded === true} />
          <Row label="cmpStatus" value={ping?.cmpStatus ?? "(none)"} />
          <Row label="displayStatus" value={ping?.displayStatus ?? tcData?.displayStatus ?? "(none)"} />
          <Row label="cmpId" value={fmt(ping?.cmpId ?? tcData?.cmpId)} />
          <Row label="cmpVersion" value={fmt(ping?.cmpVersion ?? tcData?.cmpVersion)} />
          <Row label="gdprApplies" value={fmt(ping?.gdprApplies ?? tcData?.gdprApplies)} />
          <Row label="isServiceSpecific" value={fmt(tcData?.isServiceSpecific)} />
          <Row label="gvlVersion" value={fmt(ping?.gvlVersion)} />
          <Row
            label="tcfPolicyVersion"
            value={fmt(ping?.tcfPolicyVersion ?? tcData?.tcfPolicyVersion)}
          />
          <Row label="publisherCC" value={tcData?.publisherCC ?? "(none)"} />
          <Row label="eventStatus" value={tcData?.eventStatus ?? "(none)"} />
          <Row label="apiVersion" value={ping?.apiVersion ?? "(none)"} />
        </tbody>
      </table>

      <div style={{ marginBottom: 12 }}>
        <strong>TC string</strong>{" "}
        <span style={{ fontFamily: "monospace", fontSize: 13, color: "#666" }}>
          ({tcString ? `${new TextEncoder().encode(tcString).length} bytes` : "empty"})
        </span>
        <pre
          style={{
            background: "#f0f0f0",
            padding: 8,
            border: "1px solid #ddd",
            overflow: "auto",
            margin: "4px 0 0 0",
            wordBreak: "break-all",
            whiteSpace: "pre-wrap",
          }}
        >
          {tcString || "(none)"}
        </pre>
      </div>

      <table
        style={{
          borderCollapse: "collapse",
          fontFamily: "monospace",
          fontSize: 13,
          marginBottom: 12,
        }}
      >
        <tbody>
          <Row label="purpose consents" value={purposeConsents.join(", ") || "(none)"} />
          <Row label="purpose LI" value={purposeLI.join(", ") || "(none)"} />
          <Row label="special feature opt-ins" value={specialFeatures.join(", ") || "(none)"} />
          <Row label="vendor consents" value={`${vendorConsents.length} granted`} />
          <Row label="vendor LI" value={`${vendorLI.length} granted`} />
        </tbody>
      </table>

      <strong>Vendor IDs</strong>
      <pre
        style={{
          background: "#f0f0f0",
          padding: 8,
          border: "1px solid #ddd",
          overflow: "auto",
          margin: "4px 0 0 0",
          maxHeight: 160,
        }}
      >
        {JSON.stringify(
          { consents: vendorConsents, legitimateInterests: vendorLI },
          null,
          2,
        )}
      </pre>
    </div>
  );
}

function fmt(value: boolean | number | undefined): string {
  if (value === undefined) {
    return "(none)";
  }
  return String(value);
}

function Row({
  label,
  value,
  ok,
}: {
  label: string;
  value: string;
  ok?: boolean;
}) {
  return (
    <tr>
      <td style={{ padding: "2px 16px 2px 0", color: "#666" }}>{label}</td>
      <td
        style={{
          padding: "2px 0",
          color: ok ? "green" : undefined,
          fontWeight: ok ? "bold" : "normal",
        }}
      >
        {value}
      </td>
    </tr>
  );
}
