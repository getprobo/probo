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
import { TCString } from "@iabtechlabtcf/core";
import type { BannerConfig, ConsentAction, TCFChoices } from "@probo/cookie-banner";
import { setLayoutRenderer, setTCFRuntime } from "@probo/cookie-banner";

import {
  encodeTCString,
  gdprApplies,
  requireCmpID,
  requireCmpVersion,
  type TCFGrant,
} from "./encode";
import { renderTCFLayout, wireTCFLayout } from "./layout";
import { getLastTCString, setLastTCString } from "./session";
import { disableTCFStub } from "./stub";

function tcfActive(config: BannerConfig): boolean {
  return !!config.tcf && gdprApplies(config) && !!config.tcf.gvl;
}

export function startTCF(): void {
  let cmpApi: CmpApi | undefined;
  let pending: TCFChoices | undefined;
  let lastConfig: BannerConfig | undefined;
  let active = false;
  let uiVisible = false;

  function ensureCmpApi(cmpId?: number, cmpVersion?: number): CmpApi {
    if (!cmpApi) {
      cmpApi = new CmpApi(requireCmpID(cmpId), requireCmpVersion(cmpVersion), true);
    }

    return cmpApi;
  }

  function previewPending(choices: TCFChoices): void {
    if (!active || !cmpApi || !lastConfig || !uiVisible) {
      return;
    }

    try {
      cmpApi.update(encodeTCString(lastConfig, choices), true);
    } catch {
      // Leave the last successful CmpApi state in place.
    }
  }

  setLayoutRenderer({
    render: renderTCFLayout,
    wire: wireTCFLayout,
  });

  setTCFRuntime({
    onConfig(config, existingTc) {
      lastConfig = config;
      active = tcfActive(config);
      if (!active) {
        setLastTCString(undefined);
        if (cmpApi) {
          cmpApi.update(null);
        } else {
          disableTCFStub();
        }
        return;
      }

      const api = ensureCmpApi(config.tcf?.cmp_id, config.tcf?.cmp_version);

      if (existingTc) {
        try {
          api.update(existingTc, false);
          setLastTCString(existingTc);
          return;
        } catch {
          // A stale or malformed client cookie must not suppress the banner.
        }
      }

      setLastTCString(undefined);
      api.update("", true);
    },
    onUIVisible(visible) {
      uiVisible = visible;
      if (!active || !cmpApi) {
        return;
      }

      cmpApi.update(getLastTCString() ?? "", visible);
    },
    setPendingChoices(choices) {
      pending = choices;
      previewPending(choices);
    },
    getPendingChoices() {
      return pending;
    },
    decodeChoices(tc) {
      return decodeTCChoices(tc);
    },
    onConsent(action, config) {
      lastConfig = config;
      active = tcfActive(config);
      if (!active) {
        setLastTCString(undefined);
        if (cmpApi) {
          cmpApi.update(null);
        } else {
          disableTCFStub();
        }
        return undefined;
      }

      const api = ensureCmpApi(config.tcf?.cmp_id, config.tcf?.cmp_version);

      const grant = grantForAction(action, pending);
      pending = undefined;
      const tc = encodeTCString(config, grant);
      setLastTCString(tc);
      try {
        api.update(tc, false);
      } catch {
        // Persistence must not depend on CmpApi accepting the model.
      }
      return tc;
    },
  });
}

export function grantForAction(
  action: ConsentAction,
  pending: TCFChoices | undefined,
): TCFGrant {
  if (action === "ACCEPT_ALL" || action === "ACKNOWLEDGE") {
    return "all";
  }
  if (action === "CUSTOMIZE" && pending) {
    return pending;
  }
  return "none";
}

function decodeTCChoices(tc: string): TCFChoices | null {
  try {
    const model = TCString.decode(tc);
    return {
      purposeConsents: vectorIds(model.purposeConsents),
      purposeLegitimateInterests: vectorIds(model.purposeLegitimateInterests),
      vendorConsents: vectorIds(model.vendorConsents),
      vendorLegitimateInterests: vectorIds(model.vendorLegitimateInterests),
      specialFeatureOptins: vectorIds(model.specialFeatureOptins),
    };
  } catch {
    return null;
  }
}

function vectorIds(vector: { maxId: number; has(id: number): boolean }): number[] {
  const ids: number[] = [];
  for (let id = 1; id <= vector.maxId; id++) {
    if (vector.has(id)) {
      ids.push(id);
    }
  }
  return ids;
}
