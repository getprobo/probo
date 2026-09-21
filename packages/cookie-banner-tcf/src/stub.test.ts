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

import { afterEach, describe, expect, it, vi } from "vitest";

import { disableTCFStub, installTCFStub } from "./stub";

type TCFAPI = ((...args: unknown[]) => unknown) & { q?: unknown[][] };
type TCFAPICallback = (data: unknown, success: boolean) => void;

function stubWindow(): Window & { __tcfapi?: TCFAPI } {
  const w = {} as Window & { __tcfapi?: TCFAPI };
  vi.stubGlobal("window", w);
  return w;
}

describe("installTCFStub", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("answers ping immediately and queues other commands", () => {
    const w = stubWindow();
    installTCFStub();

    expect(typeof w.__tcfapi).toBe("function");

    let ping: { cmpStatus?: string; cmpId?: number; cmpVersion?: number } | undefined;
    w.__tcfapi?.("ping", 2, (data: unknown) => {
      ping = data as typeof ping;
    });
    expect(ping).toEqual({
      gdprApplies: true,
      cmpLoaded: false,
      cmpStatus: "stub",
      displayStatus: "hidden",
      apiVersion: "2.2",
    });

    const listener = (): void => {};
    w.__tcfapi?.("addEventListener", 2, listener);
    expect(w.__tcfapi?.q).toEqual([["addEventListener", 2, listener]]);
    expect(w.__tcfapi?.()).toEqual([["addEventListener", 2, listener]]);
  });

  it("publishes a disabled ping without constructing CmpApi", () => {
    const w = stubWindow();
    installTCFStub();
    disableTCFStub();

    let ping: Record<string, unknown> | undefined;
    w.__tcfapi?.("ping", 2, (data: unknown) => {
      ping = data as Record<string, unknown>;
    });
    expect(ping).toEqual({
      gdprApplies: false,
      cmpLoaded: true,
      cmpStatus: "loaded",
      displayStatus: "disabled",
      apiVersion: "2.2",
    });
  });

  it("does not replace an existing __tcfapi", () => {
    const existing = Object.assign(() => {}, { q: [["kept"]] });
    const w = stubWindow();
    w.__tcfapi = existing as TCFAPI;

    installTCFStub();
    expect(w.__tcfapi).toBe(existing);
  });

  it("forwards locator messages to the current __tcfapi", () => {
    const listeners: Array<(event: MessageEvent) => void> = [];
    const postMessage = vi.fn();
    const w = {
      addEventListener: (_type: string, handler: (event: MessageEvent) => void) => {
        listeners.push(handler);
      },
      document: {
        body: { appendChild: vi.fn() },
        createElement: () => ({ style: { cssText: "" } }),
      },
      frames: {},
    } as unknown as Window & { __tcfapi?: TCFAPI };
    vi.stubGlobal("window", w);

    installTCFStub();
    const replacement = vi.fn();
    w.__tcfapi = replacement as TCFAPI;

    listeners[0]?.({
      data: { __tcfapiCall: { command: "ping", version: 2, callId: "1" } },
      source: { postMessage },
    } as unknown as MessageEvent);

    expect(replacement).toHaveBeenCalledWith(
      "ping",
      2,
      expect.any(Function),
      undefined,
    );

    const callback = replacement.mock.calls[0]?.[2] as TCFAPICallback;
    callback({ cmpStatus: "loaded" }, true);
    expect(postMessage).toHaveBeenCalledWith(
      {
        __tcfapiReturn: {
          returnValue: { cmpStatus: "loaded" },
          success: true,
          callId: "1",
        },
      },
      "*",
    );
  });
});
