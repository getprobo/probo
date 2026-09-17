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

import { TCString } from "@iabtechlabtcf/core";
import type {
  BannerConfig,
  LayoutHost,
  TCFChoices,
  TCFGVL,
  TCFGVLVendor,
} from "@probo/cookie-banner";
import {
  BRANDING,
  CLOSE_ICON,
  esc,
  floatingCard,
  getTCFRuntime,
} from "@probo/cookie-banner";

import { gdprApplies } from "./encode";
import { getLastTCString } from "./session";

interface Named {
  id: number;
  name: string;
  description?: string;
}

export function renderTCFLayout(config: BannerConfig, position: string): string | null {
  const gvl = config.tcf?.gvl;
  if (!gvl || !gdprApplies(config)) {
    return null;
  }

  return renderBanner(gvl, position) + renderPanel(config, gvl, position);
}

export function wireTCFLayout(root: LayoutHost, host: ShadowRoot): void {
  if (!root.bannerConfig.tcf?.gvl) {
    return;
  }

  applyLastTC(host);

  host.addEventListener("probo-state", (e: Event) => {
    if ((e as CustomEvent).detail.state === "panel") {
      applyLastTC(host);
    }
  });

  host.addEventListener(
    "click",
    (e: Event) => {
      const target = e.target as Element | null;
      if (!target?.closest?.("probo-save-button")) {
        return;
      }
      getTCFRuntime()?.setPendingChoices?.(
        collectChoices(host, root.bannerConfig),
      );
    },
    true,
  );

  host.querySelector("[data-action=open-vendors]")?.addEventListener("click", () => {
    root.setState("panel");
    host.getElementById("probo-tcf-vendors")?.scrollIntoView({ block: "start" });
  });
}

function renderBanner(gvl: TCFGVL, position: string): string {
  const purposes = namedList(gvl.purposes);
  const specialFeatures = usedSpecialFeatures(gvl);
  const vendorCount = Object.keys(gvl.vendors).length;
  const partnerLabel = vendorCount === 1 ? "1 partner" : `${vendorCount} partners`;

  const extras = [
    `<p class="description">This site stores and/or accesses information on a device and processes personal data.</p>`,
    purposes.length
      ? `<p class="description">Purposes: ${esc(purposes.map((p) => p.name).join(", "))}.</p>`
      : "",
    specialFeatures.length
      ? `<p class="description">Special features: ${esc(specialFeatures.map((f) => f.name).join(", "))}.</p>`
      : "",
    `<p class="description">We work with ${esc(partnerLabel)}. <button type="button" class="btn-link" data-action="open-vendors">View partners</button></p>`,
  ].join("");

  return `
    <probo-banner>
      ${floatingCard(
        position,
        { labelledby: "probo-banner-title", describedby: "probo-banner-desc" },
        `
        <p class="title" id="probo-banner-title" data-text="banner_title"></p>
        <p class="description" id="probo-banner-desc" data-text="banner_description"></p>
        ${extras}
        <div class="buttons">
          <probo-accept-button><button class="btn btn-primary" data-text="button_accept_all"></button></probo-accept-button>
          <probo-reject-button><button class="btn" data-text="button_reject_all"></button></probo-reject-button>
          <probo-customize-button><button class="btn btn-link" data-text="button_customize"></button></probo-customize-button>
        </div>
        ${BRANDING}`,
      )}
    </probo-banner>`;
}

function renderPanel(config: BannerConfig, gvl: TCFGVL, position: string): string {
  const purposes = namedList(gvl.purposes);
  const specialFeatures = usedSpecialFeatures(gvl);
  const vendors = Object.values(gvl.vendors).sort((a, b) => a.id - b.id);
  const purposeNames = new Map(purposes.map((p) => [p.id, p.name]));
  const days = config.consent_expiry_days;

  return `
    <probo-preference-panel>
      ${floatingCard(
        position,
        { labelledby: "probo-panel-title", describedby: "probo-panel-desc" },
        `
        <div class="panel-header">
          <div class="panel-header-title">
            <p class="title" id="probo-panel-title" style="margin:0" data-text="panel_title"></p>
            <button class="panel-close" data-action="back" data-aria-text="aria_close">
              ${CLOSE_ICON}
            </button>
          </div>
          <p class="description" id="probo-panel-desc" data-text="panel_description"></p>
        </div>
        <div class="panel-body">
          <div class="section-title">Purposes</div>
          ${purposes.map((p) => row(p.name, p.description, toggle("purpose-consent", p.id, "Consent"))).join("")}
          ${
            specialFeatures.length
              ? `<div class="section-title">Special features</div>${specialFeatures
                  .map((f) =>
                    row(f.name, f.description, toggle("special-feature", f.id, "Opt-in")),
                  )
                  .join("")}`
              : ""
          }
          <div class="section-title" id="probo-tcf-vendors">Partners</div>
          ${vendors.map((v) => vendorRow(v, purposeNames)).join("")}
          <div class="section-title">Storage</div>
          <p class="description" style="padding: 0 24px 16px">Your choices are stored in the probo_consent cookie for ${days} days.</p>
        </div>
        <div class="footer">
          <div class="buttons">
            <probo-accept-button><button class="btn btn-primary" data-text="button_accept_all"></button></probo-accept-button>
            <probo-reject-button><button class="btn" data-text="button_reject_all"></button></probo-reject-button>
            <probo-save-button>
              <button class="btn btn-link" style="flex:1" data-text="button_save"></button>
            </probo-save-button>
          </div>
          ${BRANDING}
        </div>`,
      )}
    </probo-preference-panel>`;
}

function vendorRow(vendor: TCFGVLVendor, purposeNames: Map<number, string>): string {
  const controls = [
    toggle("vendor-consent", vendor.id, "Consent"),
    vendor.legIntPurposes?.length
      ? toggle("vendor-li", vendor.id, "Legitimate interest")
      : "",
  ].join("");

  return row(vendor.name, vendorDescription(vendor, purposeNames), `<div class="toggle-group">${controls}</div>`);
}

function vendorDescription(
  vendor: TCFGVLVendor,
  purposeNames: Map<number, string>,
): string | undefined {
  const ids = [...new Set([...(vendor.purposes ?? []), ...(vendor.legIntPurposes ?? [])])];
  const names = ids
    .map((id) => purposeNames.get(id))
    .filter((name): name is string => !!name);
  const bits: string[] = [];
  if (names.length) {
    bits.push(names.join(", "));
  }
  if (vendor.policyUrl) {
    bits.push(`Privacy policy: ${vendor.policyUrl}`);
  }
  return bits.length ? bits.join(". ") : undefined;
}

function row(name: string, description: string | undefined, controls: string): string {
  return `<div class="category-header">
    <div class="category-info">
      <div class="category-name">${esc(name)}</div>
      ${description ? `<div class="category-description">${esc(description)}</div>` : ""}
    </div>
    ${controls}
  </div>`;
}

function toggle(kind: string, id: number, label: string): string {
  return `<label class="toggle" title="${esc(label)}">
    <input type="checkbox" data-tcf="${esc(kind)}" data-id="${id}">
    <span class="toggle-track"></span>
  </label>`;
}

function namedList(record: Record<string, unknown> | undefined): Named[] {
  if (!record) {
    return [];
  }

  const out: Named[] = [];
  for (const value of Object.values(record)) {
    if (!value || typeof value !== "object") {
      continue;
    }
    const rec = value as { id?: unknown; name?: unknown; description?: unknown };
    if (typeof rec.id !== "number" || typeof rec.name !== "string") {
      continue;
    }
    out.push({
      id: rec.id,
      name: rec.name,
      description: typeof rec.description === "string" ? rec.description : undefined,
    });
  }
  return out.sort((a, b) => a.id - b.id);
}

function usedSpecialFeatures(gvl: TCFGVL): Named[] {
  const used = new Set<number>();
  for (const vendor of Object.values(gvl.vendors)) {
    for (const id of vendor.specialFeatures ?? []) {
      used.add(id);
    }
  }
  return namedList(gvl.specialFeatures).filter((f) => used.has(f.id));
}

function collectChoices(host: ParentNode, config: BannerConfig): TCFChoices {
  const vendorLegitimateInterests = checkedIds(host, '[data-tcf="vendor-li"]');
  return {
    purposeConsents: checkedIds(host, '[data-tcf="purpose-consent"]'),
    purposeLegitimateInterests: purposeLIFromVendors(config, vendorLegitimateInterests),
    vendorConsents: checkedIds(host, '[data-tcf="vendor-consent"]'),
    vendorLegitimateInterests,
    specialFeatureOptins: checkedIds(host, '[data-tcf="special-feature"]'),
  };
}

function purposeLIFromVendors(config: BannerConfig, vendorIds: number[]): number[] {
  const vendors = config.tcf?.gvl?.vendors;
  if (!vendors) {
    return [];
  }

  const ids = new Set<number>();
  for (const vendorId of vendorIds) {
    const vendor = vendors[String(vendorId)];
    for (const purposeId of vendor?.legIntPurposes ?? []) {
      ids.add(purposeId);
    }
  }
  return [...ids].sort((a, b) => a - b);
}

function checkedIds(host: ParentNode, selector: string): number[] {
  return [...host.querySelectorAll<HTMLInputElement>(selector)]
    .filter((el) => el.checked)
    .map((el) => Number(el.dataset.id))
    .filter((id) => Number.isInteger(id) && id > 0);
}

function applyLastTC(host: ParentNode): void {
  const tc = getLastTCString();
  if (!tc) {
    return;
  }

  try {
    const model = TCString.decode(tc);
    for (const input of host.querySelectorAll<HTMLInputElement>("[data-tcf][data-id]")) {
      const id = Number(input.dataset.id);
      switch (input.dataset.tcf) {
        case "purpose-consent":
          input.checked = model.purposeConsents.has(id);
          break;
        case "special-feature":
          input.checked = model.specialFeatureOptins.has(id);
          break;
        case "vendor-consent":
          input.checked = model.vendorConsents.has(id);
          break;
        case "vendor-li":
          input.checked = model.vendorLegitimateInterests.has(id);
          break;
        default:
          break;
      }
    }
  } catch {
    // Leave the panel at its default (all off) if the stored string is unreadable.
  }
}
