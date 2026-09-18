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

type TCFAPICallback = (data: unknown, success: boolean) => void;

type TCFAPIStub = ((...args: unknown[]) => void) & {
  q: unknown[][];
};

type LocatorWindow = Window & {
  __tcfapi?: TCFAPIStub;
  frames?: { __tcfapiLocator?: Window };
};

const stubPing = {
  gdprApplies: true,
  cmpLoaded: false,
  cmpStatus: "stub",
  displayStatus: "hidden",
  // CMP JS API version. IAB did not bump this for TCF 2.3; the
  // encoded string is 2.3 via the disclosed-vendors segment.
  apiVersion: "2.2",
};

const disabledPing = {
  gdprApplies: false,
  cmpLoaded: true,
  cmpStatus: "loaded",
  displayStatus: "disabled",
  apiVersion: "2.2",
};

let disabled = false;

export function installTCFStub(): void {
  const w = window as LocatorWindow;

  if (typeof w.__tcfapi === "function") {
    return;
  }

  disabled = false;
  const queue: unknown[][] = [];
  const stub: TCFAPIStub = (...args: unknown[]): void => {
    const command = args[0];
    const callback = args[2];

    if (command === "ping" && typeof callback === "function") {
      (callback as TCFAPICallback)(disabled ? disabledPing : stubPing, true);
      return;
    }

    queue.push(args);
  };
  stub.q = queue;
  w.__tcfapi = stub;
  installLocator(w, stub);
}

export function disableTCFStub(): void {
  if (typeof window === "undefined") {
    return;
  }

  const stub = (window as LocatorWindow).__tcfapi;
  if (typeof stub !== "function" || !Array.isArray(stub.q)) {
    return;
  }

  disabled = true;
}

function installLocator(w: LocatorWindow, stub: TCFAPIStub): void {
  if (w.frames?.__tcfapiLocator) {
    return;
  }

  const addFrame = (): void => {
    const doc = w.document;
    if (!doc?.body) {
      w.setTimeout?.(addFrame, 5);
      return;
    }

    if (w.frames?.__tcfapiLocator) {
      return;
    }

    const iframe = doc.createElement("iframe");
    iframe.style.cssText = "display:none";
    iframe.name = "__tcfapiLocator";
    iframe.title = "__tcfapiLocator";
    doc.body.appendChild(iframe);
  };

  addFrame();

  w.addEventListener?.("message", (event: MessageEvent) => {
    let payload: {
      command?: unknown;
      version?: unknown;
      callId?: unknown;
      parameter?: unknown;
    } | undefined;

    try {
      const data = typeof event.data === "string" ? JSON.parse(event.data) : event.data;
      payload = data?.__tcfapiCall;
    } catch {
      return;
    }

    if (!payload || typeof payload.command !== "string") {
      return;
    }

    stub(
      payload.command,
      payload.version,
      (returnValue: unknown, success: boolean) => {
        const returnMsg = {
          __tcfapiReturn: {
            returnValue,
            success,
            callId: payload.callId,
          },
        };
        const source = event.source as Window | null;
        if (!source || typeof source.postMessage !== "function") {
          return;
        }
        source.postMessage(
          typeof event.data === "string" ? JSON.stringify(returnMsg) : returnMsg,
          "*",
        );
      },
      payload.parameter,
    );
  });
}
