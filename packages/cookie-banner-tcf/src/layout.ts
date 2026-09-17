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
  interpolate,
} from "@probo/cookie-banner";

import { gdprApplies } from "./encode";
import { getLastTCString } from "./session";

interface Named {
  id: number;
  name: string;
  description?: string;
  illustrations?: string[];
}

interface VendorURL {
  privacy?: string;
  legIntClaim?: string;
}

type PanelVendor = TCFGVLVendor & {
  urls?: VendorURL[];
  dataDeclaration?: number[];
};

interface Stack extends Named {
  purposes: number[];
  specialFeatures: number[];
}

export function renderTCFLayout(config: BannerConfig, position: string): string | null {
  const gvl = config.tcf?.gvl;
  if (!gvl || !gdprApplies(config)) {
    return null;
  }

  return renderBanner(config, gvl, position) + renderPanel(config, gvl, position);
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
        collectChoices(host),
      );
    },
    true,
  );

  host.querySelector("[data-action=open-vendors]")?.addEventListener("click", () => {
    root.setState("panel");
    host.getElementById("probo-tcf-vendors")?.scrollIntoView({ block: "start" });
  });
}

function renderBanner(config: BannerConfig, gvl: TCFGVL, position: string): string {
  const purposes = namedList(gvl.purposes);
  const specialFeatures = usedSpecialFeatures(gvl);
  const vendorCount = Object.keys(gvl.vendors).length;
  const partnerLabel = interpolate(
    text(
      config,
      vendorCount === 1 ? "tcf_partner_one" : "tcf_partners",
      vendorCount === 1 ? "{{count}} partner" : "{{count}} partners",
    ),
    { count: String(vendorCount) },
  );

  const extras = [
    `<p class="description" data-text="tcf_disclosure_store">This site stores and/or accesses information on a device and processes personal data.</p>`,
    purposes.length
      ? `<p class="description"><span data-text="tcf_label_purposes">Purposes</span>: ${esc(purposes.map((p) => p.name).join(", "))}.</p>`
      : "",
    specialFeatures.length
      ? `<p class="description"><span data-text="tcf_label_special_features">Special features</span>: ${esc(specialFeatures.map((f) => f.name).join(", "))}.</p>`
      : "",
    `<p class="description">${esc(text(config, "tcf_disclosure_partners", "We work with {{partners}}.", { partners: partnerLabel }))} <button type="button" class="btn-link" data-action="open-vendors" data-text="tcf_view_partners">View partners</button></p>`,
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
  const liPurposeIDs = purposeLIIds(gvl);
  const { stacks, ungrouped } = groupPurposes(gvl, purposes);
  const storageCopy = interpolate(
    text(config, "tcf_storage", "Your choices are stored in the probo_consent cookie for {{days}} days."),
    { days: String(config.consent_expiry_days) },
  );

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
          ${stacks.map((stack) => stackSection(stack, purposes, liPurposeIDs)).join("")}
          ${
            ungrouped.length
              ? `<div class="section-title" data-text="tcf_section_purposes">Purposes</div>${ungrouped
                  .map((p) => purposeRow(p, liPurposeIDs.has(p.id)))
                  .join("")}`
              : ""
          }
          ${disclosureSection("Special purposes", usedSpecialPurposes(gvl))}
          ${disclosureSection("Features", usedFeatures(gvl))}
          ${
            specialFeatures.length
              ? `<div class="section-title" data-text="tcf_section_special_features">Special features</div>${specialFeatures
                  .map((f) =>
                    row(f.name, namedBody(f), toggle("special-feature", f.id, "Opt-in")),
                  )
                  .join("")}`
              : ""
          }
          ${disclosureSection("Data categories", usedDataCategories(gvl))}
          <div class="section-title" id="probo-tcf-vendors" data-text="tcf_section_partners">Partners</div>
          ${vendors.map((v) => vendorRow(v, purposeNames)).join("")}
          <div class="section-title" data-text="tcf_section_storage">Storage</div>
          <p class="description" style="padding: 0 24px 16px">${esc(storageCopy)}</p>
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

function stackSection(stack: Stack, purposes: Named[], liPurposeIDs: Set<number>): string {
  const purposeByID = new Map(purposes.map((p) => [p.id, p]));
  const rows = stack.purposes
    .map((id) => purposeByID.get(id))
    .filter((p): p is Named => !!p)
    .map((p) => purposeRow(p, liPurposeIDs.has(p.id)))
    .join("");

  if (!rows) {
    return "";
  }

  return `<div class="section-title">${esc(stack.name)}</div>${
    stack.description ? `<p class="description" style="padding: 0 24px 8px">${esc(stack.description)}</p>` : ""
  }${rows}`;
}

function purposeRow(purpose: Named, showLI: boolean): string {
  const controls = [
    toggle("purpose-consent", purpose.id, "Consent"),
    showLI ? toggle("purpose-li", purpose.id, "Legitimate interest") : "",
  ].join("");

  return row(purpose.name, namedBody(purpose), `<div class="toggle-group">${controls}</div>`);
}

function vendorRow(vendor: PanelVendor, purposeNames: Map<number, string>): string {
  const controls = [
    toggle("vendor-consent", vendor.id, "Consent"),
    vendor.legIntPurposes?.length
      ? toggle("vendor-li", vendor.id, "Legitimate interest")
      : "",
  ].join("");

  return row(vendor.name, vendorBody(vendor, purposeNames), `<div class="toggle-group">${controls}</div>`);
}

function disclosureSection(title: string, items: Named[]): string {
  if (!items.length) {
    return "";
  }

  return `<div class="section-title">${esc(title)}</div>${items
    .map((item) => row(item.name, namedBody(item), ""))
    .join("")}`;
}

function namedBody(item: Named): string | undefined {
  const parts = [item.description, ...(item.illustrations ?? [])]
    .filter((part): part is string => !!part)
    .map((part) => esc(part));
  return parts.length ? parts.join(" ") : undefined;
}

function vendorBody(vendor: PanelVendor, purposeNames: Map<number, string>): string | undefined {
  const ids = [...new Set([...(vendor.purposes ?? []), ...(vendor.legIntPurposes ?? [])])];
  const names = ids
    .map((id) => purposeNames.get(id))
    .filter((name): name is string => !!name);
  const bits: string[] = [];
  if (names.length) {
    bits.push(esc(names.join(", ")));
  }
  const storage = vendorStorage(vendor);
  if (storage) {
    bits.push(esc(storage));
  }
  const links = vendorLinks(vendor);
  if (links) {
    bits.push(links);
  }
  return bits.length ? bits.join(". ") : undefined;
}

function vendorStorage(vendor: PanelVendor): string | undefined {
  const bits: string[] = [];
  if (vendor.usesCookies) {
    const age = formatCookieMaxAge(vendor.cookieMaxAgeSeconds);
    bits.push(age ? `Cookies (up to ${age})` : "Cookies");
  }
  if (vendor.usesNonCookieAccess) {
    bits.push("Non-cookie storage");
  }
  return bits.length ? bits.join(". ") : undefined;
}

function formatCookieMaxAge(seconds: number | null | undefined): string | undefined {
  if (seconds == null || !Number.isFinite(seconds) || seconds < 0) {
    return undefined;
  }
  if (seconds === 0) {
    return "session";
  }
  const days = Math.round(seconds / 86400);
  if (days >= 1) {
    return days === 1 ? "1 day" : `${days} days`;
  }
  const hours = Math.round(seconds / 3600);
  if (hours >= 1) {
    return hours === 1 ? "1 hour" : `${hours} hours`;
  }
  return `${Math.round(seconds)} seconds`;
}

function vendorLinks(vendor: PanelVendor): string | undefined {
  const first = vendor.urls?.[0];
  const privacy = httpUrl(first?.privacy ?? vendor.policyUrl);
  const claim = httpUrl(first?.legIntClaim);
  const bits: string[] = [];
  if (privacy) {
    bits.push(policyAnchor(privacy, "Privacy policy"));
  }
  if (claim) {
    bits.push(policyAnchor(claim, "Legitimate interest"));
  }
  return bits.length ? bits.join(". ") : undefined;
}

function policyAnchor(href: string, label: string): string {
  return `<a class="btn-link" href="${esc(href)}" target="_blank" rel="noopener noreferrer">${esc(label)}</a>`;
}

function httpUrl(value: string | undefined): string | undefined {
  if (!value) {
    return undefined;
  }

  try {
    const url = new URL(value);
    if (url.protocol === "http:" || url.protocol === "https:") {
      return url.href;
    }
  } catch {
    // Vendor-declared URLs that are not http(s) stay off the page.
  }

  return undefined;
}

function row(name: string, descriptionHtml: string | undefined, controls: string): string {
  return `<div class="category-header">
    <div class="category-info">
      <div class="category-name">${esc(name)}</div>
      ${descriptionHtml ? `<div class="category-description">${descriptionHtml}</div>` : ""}
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
    const rec = value as {
      id?: unknown;
      name?: unknown;
      description?: unknown;
      illustrations?: unknown;
    };
    if (typeof rec.id !== "number" || typeof rec.name !== "string") {
      continue;
    }
    out.push({
      id: rec.id,
      name: rec.name,
      description: typeof rec.description === "string" ? rec.description : undefined,
      illustrations: stringList(rec.illustrations),
    });
  }
  return out.sort((a, b) => a.id - b.id);
}

function stackList(record: Record<string, unknown> | undefined): Stack[] {
  if (!record) {
    return [];
  }

  const out: Stack[] = [];
  for (const value of Object.values(record)) {
    if (!value || typeof value !== "object") {
      continue;
    }
    const rec = value as {
      id?: unknown;
      name?: unknown;
      description?: unknown;
      purposes?: unknown;
      specialFeatures?: unknown;
    };
    if (typeof rec.id !== "number" || typeof rec.name !== "string") {
      continue;
    }
    out.push({
      id: rec.id,
      name: rec.name,
      description: typeof rec.description === "string" ? rec.description : undefined,
      purposes: numberIDs(rec.purposes),
      specialFeatures: numberIDs(rec.specialFeatures),
    });
  }
  return out.sort((a, b) => a.id - b.id);
}

function numberIDs(value: unknown): number[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value.filter((id): id is number => typeof id === "number" && Number.isInteger(id) && id > 0);
}

function stringList(value: unknown): string[] | undefined {
  if (!Array.isArray(value)) {
    return undefined;
  }
  const out = value.filter((item): item is string => typeof item === "string" && item.length > 0);
  return out.length ? out : undefined;
}

function usedNamed(
  catalog: Record<string, unknown> | undefined,
  idsFromVendor: (vendor: PanelVendor) => number[] | undefined,
  gvl: TCFGVL,
): Named[] {
  const used = new Set<number>();
  for (const vendor of Object.values(gvl.vendors) as PanelVendor[]) {
    for (const id of idsFromVendor(vendor) ?? []) {
      used.add(id);
    }
  }
  return namedList(catalog).filter((item) => used.has(item.id));
}

function usedSpecialFeatures(gvl: TCFGVL): Named[] {
  return usedNamed(gvl.specialFeatures, (vendor) => vendor.specialFeatures, gvl);
}

function usedSpecialPurposes(gvl: TCFGVL): Named[] {
  return usedNamed(gvl.specialPurposes, (vendor) => vendor.specialPurposes, gvl);
}

function usedFeatures(gvl: TCFGVL): Named[] {
  return usedNamed(gvl.features, (vendor) => vendor.features, gvl);
}

function usedDataCategories(gvl: TCFGVL): Named[] {
  return usedNamed(gvl.dataCategories, (vendor) => vendor.dataDeclaration, gvl);
}

function purposeLIIds(gvl: TCFGVL): Set<number> {
  const ids = new Set<number>();
  for (const vendor of Object.values(gvl.vendors)) {
    for (const id of vendor.legIntPurposes ?? []) {
      ids.add(id);
    }
  }
  return ids;
}

function groupPurposes(gvl: TCFGVL, purposes: Named[]): { stacks: Stack[]; ungrouped: Named[] } {
  const assigned = new Set<number>();
  const stacks: Stack[] = [];

  for (const stack of stackList(gvl.stacks)) {
    const purposeIDs = stack.purposes.filter((id) => purposes.some((p) => p.id === id) && !assigned.has(id));
    if (!purposeIDs.length) {
      continue;
    }
    for (const id of purposeIDs) {
      assigned.add(id);
    }
    stacks.push({ ...stack, purposes: purposeIDs });
  }

  return {
    stacks,
    ungrouped: purposes.filter((p) => !assigned.has(p.id)),
  };
}

function collectChoices(host: ParentNode): TCFChoices {
  return {
    purposeConsents: checkedIds(host, '[data-tcf="purpose-consent"]'),
    purposeLegitimateInterests: checkedIds(host, '[data-tcf="purpose-li"]'),
    vendorConsents: checkedIds(host, '[data-tcf="vendor-consent"]'),
    vendorLegitimateInterests: checkedIds(host, '[data-tcf="vendor-li"]'),
    specialFeatureOptins: checkedIds(host, '[data-tcf="special-feature"]'),
  };
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
        case "purpose-li":
          input.checked = model.purposeLegitimateInterests.has(id);
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

function text(
  config: BannerConfig,
  key: string,
  fallback: string,
  vars?: Record<string, string>,
): string {
  const raw = config.texts?.[key] || fallback;
  return vars ? interpolate(raw, vars) : raw;
}
