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

import { CmpApi } from "@iabtechlabtcf/cmpapi";

import type { BannerConfig, ConsentAction } from "../types";
import { TCF_CMP_ID, TCF_CMP_VERSION } from "./constants";
import { encodeTCString, gdprApplies } from "./encode";
import { setTCFRuntime } from "../addons";

function tcfActive(config: BannerConfig): boolean {
  return !!config.tcf && gdprApplies(config) && !!config.tcf.gvl;
}

function grantsTCF(action: ConsentAction): boolean {
  return action === "ACCEPT_ALL";
}

export function startTCF(): void {
  const cmpApi = new CmpApi(TCF_CMP_ID, TCF_CMP_VERSION, true);

  setTCFRuntime({
    onConfig(config, existingTc) {
      if (!tcfActive(config)) {
        cmpApi.update(null);
        return;
      }

      if (existingTc) {
        cmpApi.update(existingTc, false);
        return;
      }

      cmpApi.update("", true);
    },
    onConsent(action, config) {
      if (!tcfActive(config)) {
        cmpApi.update(null);
        return undefined;
      }

      const tc = encodeTCString(config, grantsTCF(action));
      cmpApi.update(tc, false);
      return tc;
    },
  });
}
