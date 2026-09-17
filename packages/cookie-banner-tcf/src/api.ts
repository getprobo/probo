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
import type { BannerConfig, ConsentAction, TCFChoices } from "@probo/cookie-banner";
import { setLayoutRenderer, setTCFRuntime } from "@probo/cookie-banner";

import { TCF_CMP_ID, TCF_CMP_VERSION } from "./constants";
import { encodeTCString, gdprApplies } from "./encode";
import { renderTCFLayout, wireTCFLayout } from "./layout";
import { setLastTCString } from "./session";

function tcfActive(config: BannerConfig): boolean {
  return !!config.tcf && gdprApplies(config) && !!config.tcf.gvl;
}

export function startTCF(): void {
  const cmpApi = new CmpApi(TCF_CMP_ID, TCF_CMP_VERSION, true);
  let pending: TCFChoices | undefined;

  setLayoutRenderer({
    render: renderTCFLayout,
    wire: wireTCFLayout,
  });

  setTCFRuntime({
    onConfig(config, existingTc) {
      if (!tcfActive(config)) {
        setLastTCString(undefined);
        cmpApi.update(null);
        return;
      }

      setLastTCString(existingTc);
      if (existingTc) {
        cmpApi.update(existingTc, false);
        return;
      }

      cmpApi.update("", true);
    },
    setPendingChoices(choices) {
      pending = choices;
    },
    onConsent(action, config) {
      if (!tcfActive(config)) {
        setLastTCString(undefined);
        cmpApi.update(null);
        return undefined;
      }

      const grant = grantForAction(action, pending);
      pending = undefined;
      const tc = encodeTCString(config, grant);
      setLastTCString(tc);
      cmpApi.update(tc, false);
      return tc;
    },
  });
}

function grantForAction(
  action: ConsentAction,
  pending: TCFChoices | undefined,
): boolean | TCFChoices {
  if (action === "ACCEPT_ALL") {
    return true;
  }
  if (action === "CUSTOMIZE" && pending) {
    return pending;
  }
  return false;
}
