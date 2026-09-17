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

import { TCF_CMP_ID, TCF_CMP_VERSION } from "./constants";

type TCFAPIStub = ((...args: unknown[]) => void) & { q: unknown[][] };

export function installTCFStub(): void {
  const w = window as Window & { __tcfapi?: TCFAPIStub };

  if (typeof w.__tcfapi === "function") {
    return;
  }

  const queue: unknown[][] = [];
  const stub: TCFAPIStub = function stub(...args: unknown[]): void {
    const command = args[0];
    const callback = args[2];

    if (command === "ping" && typeof callback === "function") {
      (callback as (data: unknown, success: boolean) => void)(
        {
          gdprApplies: true,
          cmpLoaded: false,
          cmpStatus: "stub",
          displayStatus: "hidden",
          apiVersion: "2.2",
          cmpId: TCF_CMP_ID,
          cmpVersion: TCF_CMP_VERSION,
        },
        true,
      );
      return;
    }

    queue.push(args);
  };
  stub.q = queue;
  w.__tcfapi = stub;
}
