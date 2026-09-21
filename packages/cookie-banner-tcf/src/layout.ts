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
import { TCF_PANEL_STYLES } from "./panel-styles";
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
  dataRetention?: {
    stdRetention?: number;
    purposes?: Record<string, number>;
    specialPurposes?: Record<string, number>;
  };
  deviceStorageDisclosureUrl?: string;
};

interface Stack extends Named {
  purposes: number[];
  specialFeatures: number[];
}

interface VendorCatalogs {
  purposes: Map<number, string>;
  specialPurposes: Map<number, string>;
  features: Map<number, string>;
  specialFeatures: Map<number, string>;
  dataCategories: Map<number, string>;
}

interface PurposeVendorCount {
  consent: number;
  li: number;
}

export function renderTCFLayout(config: BannerConfig, position: string): string | null {
  const gvl = config.tcf?.gvl;
  if (!gvl || !gdprApplies(config)) {
    return null;
  }

  return TCF_PANEL_STYLES + renderBanner(config, gvl, position) + renderPanel(config, gvl, position);
}

export function wireTCFLayout(root: LayoutHost, host: ShadowRoot): void {
  if (!root.bannerConfig.tcf?.gvl) {
    return;
  }

  applyLastTC(host);

  host.addEventListener("probo-state", (e: Event) => {
    if ((e as CustomEvent).detail.state === "panel") {
      applyLastTC(host);
      syncChoices(host);
    }
  });

  host.addEventListener("change", (e: Event) => {
    if (!isTCFChoice(e.target)) {
      return;
    }
    syncChoices(host);
  });

  host.addEventListener(
    "click",
    (e: Event) => {
      if (isClickOn(e, "probo-reject-button")) {
        persistAction(root, "REJECT_ALL", () => root.client.rejectAll?.());
        e.stopImmediatePropagation();
        return;
      }
      if (isClickOn(e, "probo-accept-button")) {
        persistAction(root, "ACCEPT_ALL", () => root.client.acceptAll?.());
        e.stopImmediatePropagation();
        return;
      }
      if (isClickOn(e, "probo-save-button")) {
        syncChoices(host);
      }
    },
    true,
  );

  host.querySelector("[data-action=open-vendors]")?.addEventListener("click", () => {
    root.setState("panel");
    host.getElementById("probo-tcf-vendors")?.scrollIntoView({ block: "start" });
  });
}

function renderBanner(config: BannerConfig, gvl: TCFGVL, position: string): string {
  const purposes = usedPurposes(gvl);
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
  const rights = [
    `<span data-text="tcf_disclosure_scope">These choices apply to this site only (service-specific).</span>`,
    `<span data-text="tcf_disclosure_withdraw">You can withdraw or change your consent at any time via Cookie settings.</span>`,
    purposeLIIds(gvl).size
      ? `<span data-text="tcf_disclosure_object">Some partners process personal data on the basis of legitimate interest. You can object to that processing.</span>`
      : "",
  ].filter(Boolean);

  const extras = `<div class="tcf-disclosures">
      <p class="description"><span data-text="tcf_disclosure_store">This site stores and/or accesses information on a device and processes personal data.</span> <span data-text="tcf_disclosure_data">Personal data processed includes unique identifiers and browsing data.</span></p>
      <p class="description">${rights.join(" ")}</p>
      ${
        purposes.length
          ? `<p class="description"><span data-text="tcf_label_purposes">Purposes</span>: ${esc(purposes.map((p) => p.name).join(", "))}.</p>`
          : ""
      }
      ${
        specialFeatures.length
          ? `<p class="description"><span data-text="tcf_label_special_features">Special features</span>: ${esc(specialFeatures.map((f) => f.name).join(", "))}.</p>`
          : ""
      }
      <p class="description">${esc(text(config, "tcf_disclosure_partners", "We work with {{partners}}.", { partners: partnerLabel }))} <button type="button" class="btn-link" data-action="open-vendors" data-text="tcf_view_partners">View partners</button></p>
    </div>`;

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
  const purposes = usedPurposes(gvl);
  const specialFeatures = usedSpecialFeatures(gvl);
  const vendors = Object.values(gvl.vendors).sort((a, b) => a.id - b.id);
  const catalogs: VendorCatalogs = {
    purposes: nameMap(namedList(gvl.purposes)),
    specialPurposes: nameMap(namedList(gvl.specialPurposes)),
    features: nameMap(namedList(gvl.features)),
    specialFeatures: nameMap(namedList(gvl.specialFeatures)),
    dataCategories: nameMap(namedList(gvl.dataCategories)),
  };
  const liPurposeIDs = purposeLIIds(gvl);
  const { stacks, ungrouped } = groupPurposes(gvl, purposes);
  const storageCopy = interpolate(
    text(config, "tcf_storage", "Your choices are stored in the probo_consent cookie for {{days}} days."),
    { days: String(config.consent_expiry_days) },
  );

  return `
    <probo-preference-panel class="tcf-panel">
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
          <p class="description" id="probo-panel-desc" data-text="tcf_panel_description">Choose which purposes and partners to allow. Consent and legitimate interest can be set separately when both apply.</p>
          ${
            liPurposeIDs.size
              ? `<p class="description"><span data-text="tcf_disclosure_data">Personal data processed includes unique identifiers and browsing data.</span> <span data-text="tcf_disclosure_scope">These choices apply to this site only (service-specific).</span></p>`
              : ""
          }
        </div>
        <div class="panel-body">
          ${purposesSection(config, stacks, ungrouped, purposes, liPurposeIDs, purposeVendorCounts(gvl))}
          ${sectionGroup(
            "Special features",
            specialFeatures
              .map((f) =>
                choiceRow(
                  f.name,
                  namedBody(f),
                  toggle("special-feature", f.id, "Opt-in", "tcf_label_optin", f.name),
                ),
              )
              .join(""),
            "tcf_section_special_features",
          )}
          ${sectionGroup(
            "Partners",
            vendors.map((v) => vendorRow(v, catalogs)).join(""),
            "tcf_section_partners",
            "probo-tcf-vendors",
          )}
          ${sectionGroup("Storage", `<p class="tcf-storage">${esc(storageCopy)}</p>`, "tcf_section_storage")}
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

function purposesSection(
  config: BannerConfig,
  stacks: Stack[],
  ungrouped: Named[],
  purposes: Named[],
  liPurposeIDs: Set<number>,
  vendorCounts: Map<number, PurposeVendorCount>,
): string {
  const grouped = stacks
    .map((stack) => stackSection(config, stack, purposes, liPurposeIDs, vendorCounts))
    .join("");
  const leftoverRows = ungrouped
    .map((purpose) =>
      purposeRow(config, purpose, liPurposeIDs.has(purpose.id), vendorCounts.get(purpose.id)),
    )
    .join("");
  const leftover = leftoverRows
    ? grouped
      ? stackBlock("Other purposes", leftoverRows, "tcf_section_other_purposes")
      : leftoverRows
    : "";
  if (!grouped && !leftover) {
    return "";
  }

  return sectionGroup(
    "Purposes",
    `${grouped}${leftover}`,
    "tcf_section_purposes",
  );
}

function sectionGroup(title: string, body: string, textKey?: string, id?: string): string {
  if (!body) {
    return "";
  }

  return `<div class="tcf-group">${sectionTitle(title, textKey, id)}<div class="tcf-group-body">${body}</div></div>`;
}

function stackSection(
  config: BannerConfig,
  stack: Stack,
  purposes: Named[],
  liPurposeIDs: Set<number>,
  vendorCounts: Map<number, PurposeVendorCount>,
): string {
  const purposeByID = new Map(purposes.map((p) => [p.id, p]));
  const rows = stack.purposes
    .map((id) => purposeByID.get(id))
    .filter((p): p is Named => !!p)
    .map((p) => purposeRow(config, p, liPurposeIDs.has(p.id), vendorCounts.get(p.id)))
    .join("");

  if (!rows) {
    return "";
  }

  return stackBlock(stack.name, rows, undefined, stack.description);
}

function stackBlock(name: string, rows: string, textKey?: string, description?: string): string {
  return `<div class="tcf-stack">${subsectionTitle(name, textKey)}${
    description ? `<p class="tcf-section-desc">${esc(description)}</p>` : ""
  }${rows}</div>`;
}

function purposeRow(
  config: BannerConfig,
  purpose: Named,
  showLI: boolean,
  counts: PurposeVendorCount | undefined,
): string {
  const controls = [
    toggle("purpose-consent", purpose.id, "Consent", "tcf_label_consent", purpose.name),
    showLI
      ? toggle("purpose-li", purpose.id, "Legitimate interest", "tcf_label_li", purpose.name)
      : "",
  ].join("");

  return choiceRow(purpose.name, purposeBody(config, purpose, counts), controls);
}

function purposeBody(
  config: BannerConfig,
  purpose: Named,
  counts: PurposeVendorCount | undefined,
): string | undefined {
  const description = namedBody(purpose);
  const vendors = purposeVendorLine(config, counts);
  if (!description && !vendors) {
    return undefined;
  }
  return `${description ?? ""}${vendors}`;
}

function purposeVendorLine(config: BannerConfig, counts: PurposeVendorCount | undefined): string {
  if (!counts) {
    return "";
  }
  const parts = [
    counts.consent > 0
      ? esc(
          purposeCountLabel(
            config,
            counts.consent,
            "tcf_purpose_consent_one",
            "tcf_purpose_consent",
            "{{count}} partner seeking consent",
            "{{count}} partners seeking consent",
          ),
        )
      : "",
    counts.li > 0
      ? esc(
          purposeCountLabel(
            config,
            counts.li,
            "tcf_purpose_li_one",
            "tcf_purpose_li",
            "{{count}} partner relying on legitimate interest",
            "{{count}} partners relying on legitimate interest",
          ),
        )
      : "",
  ].filter(Boolean);
  if (!parts.length) {
    return "";
  }
  return `<p class="tcf-purpose-vendors">${parts.join(" · ")}</p>`;
}

function purposeCountLabel(
  config: BannerConfig,
  count: number,
  oneKey: string,
  manyKey: string,
  oneFallback: string,
  manyFallback: string,
): string {
  return text(config, count === 1 ? oneKey : manyKey, count === 1 ? oneFallback : manyFallback, {
    count: String(count),
  });
}

function vendorRow(vendor: PanelVendor, catalogs: VendorCatalogs): string {
  const controls = [
    toggle("vendor-consent", vendor.id, "Consent", "tcf_label_consent", vendor.name),
    vendor.legIntPurposes?.length
      ? toggle("vendor-li", vendor.id, "Legitimate interest", "tcf_label_li", vendor.name)
      : "",
  ].join("");

  return choiceRow(vendor.name, vendorBody(vendor, catalogs), controls);
}

function namedBody(item: Named): string | undefined {
  const parts = [item.description, ...(item.illustrations ?? [])]
    .filter((part): part is string => !!part)
    .map((part) => esc(part));
  return parts.length ? parts.join(" ") : undefined;
}

function vendorBody(vendor: PanelVendor, catalogs: VendorCatalogs): string | undefined {
  const lines = [
    vendorLine("tcf_label_consent", "Consent", catalogNames(vendor.purposes, catalogs.purposes)),
    vendorLine("tcf_label_li", "Legitimate interest", catalogNames(vendor.legIntPurposes, catalogs.purposes)),
    vendorLine(
      "tcf_section_special_purposes",
      "Special purposes",
      catalogNames(vendor.specialPurposes, catalogs.specialPurposes),
      "tcf_label_always_on",
      "Always on",
    ),
    vendorLine("tcf_section_features", "Features", catalogNames(vendor.features, catalogs.features)),
    vendorLine(
      "tcf_section_special_features",
      "Special features",
      catalogNames(vendor.specialFeatures, catalogs.specialFeatures),
    ),
    vendorLine(
      "tcf_section_data_categories",
      "Data categories",
      catalogNames(vendor.dataDeclaration, catalogs.dataCategories),
    ),
  ].filter(Boolean);
  const storage = vendorStorage(vendor);
  if (storage) {
    lines.push(`<p class="tcf-vendor-line">${storage}</p>`);
  }
  const stdRetention = vendorStdRetention(vendor);
  if (stdRetention) {
    lines.push(stdRetention);
  }
  const purposeStorage = vendorPurposeStorage(vendor, catalogs);
  if (purposeStorage) {
    lines.push(purposeStorage);
  }
  const links = vendorLinks(vendor);
  if (links) {
    lines.push(`<p class="tcf-vendor-line">${links}</p>`);
  }
  return lines.length ? `<div class="tcf-vendor-details">${lines.join("")}</div>` : undefined;
}

function vendorLine(
  labelKey: string,
  fallback: string,
  names: string[],
  noteKey?: string,
  noteFallback?: string,
): string {
  if (!names.length) {
    return "";
  }
  const note =
    noteKey && noteFallback
      ? ` (<span data-text="${esc(noteKey)}">${esc(noteFallback)}</span>)`
      : "";
  return `<p class="tcf-vendor-line"><span data-text="${esc(labelKey)}">${esc(fallback)}</span>${note}: ${esc(names.join(", "))}</p>`;
}

function catalogNames(ids: number[] | undefined, names: Map<number, string>): string[] {
  return (ids ?? []).map((id) => names.get(id)).filter((name): name is string => !!name);
}

function nameMap(items: Named[]): Map<number, string> {
  return new Map(items.map((item) => [item.id, item.name]));
}

function vendorStorage(vendor: PanelVendor): string | undefined {
  const bits: string[] = [];
  if (vendor.usesCookies) {
    const age = formatCookieMaxAge(vendor.cookieMaxAgeSeconds);
    const duration = age ? `Cookies (up to ${esc(age)})` : "Cookies";
    const refresh =
      vendor.cookieRefresh === true
        ? `<span data-text="tcf_label_cookie_refresh">may be refreshed</span>`
        : vendor.cookieRefresh === false
          ? `<span data-text="tcf_label_cookie_no_refresh">not refreshed</span>`
          : "";
    bits.push(refresh ? `${duration} (${refresh})` : duration);
  }
  if (vendor.usesNonCookieAccess) {
    bits.push(`<span data-text="tcf_label_non_cookie">Non-cookie storage</span>`);
  }
  return bits.length ? bits.join(". ") : undefined;
}

function vendorStdRetention(vendor: PanelVendor): string | undefined {
  const period = formatRetentionPeriod(vendor.dataRetention?.stdRetention);
  if (!period) {
    return undefined;
  }
  return `<p class="tcf-vendor-line"><span data-text="tcf_label_std_retention">Standard retention</span>: ${esc(period)}</p>`;
}

function vendorPurposeStorage(vendor: PanelVendor, catalogs: VendorCatalogs): string | undefined {
  const items = [
    ...retentionItems(vendor.dataRetention?.purposes, catalogs.purposes),
    ...retentionItems(vendor.dataRetention?.specialPurposes, catalogs.specialPurposes),
  ];
  if (!items.length) {
    return undefined;
  }
  return `<p class="tcf-vendor-line"><span data-text="tcf_label_purpose_storage">Purpose-specific storage</span>: ${items.join(", ")}</p>`;
}

function retentionItems(
  record: Record<string, number> | undefined,
  names: Map<number, string>,
): string[] {
  if (!record) {
    return [];
  }
  return Object.entries(record)
    .map(([id, days]) => {
      const name = names.get(Number(id));
      const period = formatRetentionPeriod(days);
      if (!name || !period) {
        return "";
      }
      return `${esc(name)} (${esc(period)})`;
    })
    .filter(Boolean);
}

function formatRetentionPeriod(days: number | undefined): string | undefined {
  if (days == null || !Number.isFinite(days) || days < 0) {
    return undefined;
  }
  if (days === 0) {
    return "session";
  }
  return formatRetentionDays(days);
}

function formatRetentionDays(days: number): string {
  const rounded = Math.round(days);
  return rounded === 1 ? "1 day" : `${rounded} days`;
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
  const disclosure = httpUrl(vendor.deviceStorageDisclosureUrl);
  const bits: string[] = [];
  if (privacy) {
    bits.push(policyAnchor(privacy, "Privacy policy"));
  }
  if (claim) {
    bits.push(policyAnchor(claim, "Legitimate interest"));
  }
  if (disclosure) {
    bits.push(policyAnchor(disclosure, "Device storage details", "tcf_label_device_storage"));
  }
  return bits.length ? bits.join(". ") : undefined;
}

function policyAnchor(href: string, label: string, textKey?: string): string {
  const textAttr = textKey ? ` data-text="${esc(textKey)}"` : "";
  return `<a class="btn-link" href="${esc(href)}" target="_blank" rel="noopener noreferrer"${textAttr}>${esc(label)}</a>`;
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

function sectionTitle(title: string, textKey?: string, id?: string): string {
  const textAttr = textKey ? ` data-text="${esc(textKey)}"` : "";
  const idAttr = id ? ` id="${esc(id)}"` : "";
  return `<div class="tcf-section"${idAttr}${textAttr}>${esc(title)}</div>`;
}

function subsectionTitle(title: string, textKey?: string): string {
  const textAttr = textKey ? ` data-text="${esc(textKey)}"` : "";
  return `<div class="tcf-subsection"${textAttr}>${esc(title)}</div>`;
}

function choiceRow(
  name: string,
  descriptionHtml: string | undefined,
  controls: string,
): string {
  return tcfRow(name, descriptionHtml, `<div class="toggle-group">${controls}</div>`);
}

function tcfRow(
  name: string,
  descriptionHtml: string | undefined,
  trailing: string,
): string {
  return `<div class="tcf-row tcf-row-choice">
    <div class="tcf-row-id" aria-hidden="true"></div>
    <div class="category-info">
      <div class="category-name">${esc(name)}</div>
      ${descriptionHtml ? `<div class="category-description">${descriptionHtml}</div>` : ""}
    </div>
    ${trailing}
  </div>`;
}

function toggle(kind: string, id: number, control: string, textKey: string, name: string): string {
  return `<label class="tcf-control">
    <span class="toggle">
      <input type="checkbox" data-tcf="${esc(kind)}" data-id="${id}" aria-label="${esc(`${control}: ${name}`)}">
      <span class="toggle-track"></span>
    </span>
    <span class="tcf-control-label" data-text="${esc(textKey)}">${esc(control)}</span>
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

function usedPurposes(gvl: TCFGVL): Named[] {
  return usedNamed(
    gvl.purposes,
    (vendor) => [...(vendor.purposes ?? []), ...(vendor.legIntPurposes ?? [])],
    gvl,
  );
}

function usedSpecialFeatures(gvl: TCFGVL): Named[] {
  return usedNamed(gvl.specialFeatures, (vendor) => vendor.specialFeatures, gvl);
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

function purposeVendorCounts(gvl: TCFGVL): Map<number, PurposeVendorCount> {
  const counts = new Map<number, PurposeVendorCount>();
  const bump = (id: number, key: keyof PurposeVendorCount) => {
    const current = counts.get(id) ?? { consent: 0, li: 0 };
    current[key] += 1;
    counts.set(id, current);
  };
  for (const vendor of Object.values(gvl.vendors)) {
    for (const id of vendor.purposes ?? []) {
      bump(id, "consent");
    }
    for (const id of vendor.legIntPurposes ?? []) {
      bump(id, "li");
    }
  }
  return counts;
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

function syncChoices(host: ParentNode): void {
  getTCFRuntime()?.setPendingChoices?.(collectChoices(host));
}

function persistAction(
  root: LayoutHost,
  action: "ACCEPT_ALL" | "REJECT_ALL",
  run: () => void,
): void {
  run();
  root.setState("hidden");
  root.dispatchEvent(
    new CustomEvent("probo-consent", {
      bubbles: true,
      composed: true,
      detail: { action },
    }),
  );
}

function isTCFChoice(target: EventTarget | null): target is HTMLInputElement {
  return target instanceof HTMLInputElement && typeof target.dataset.tcf === "string";
}

function isClickOn(e: Event, localName: string): boolean {
  if (typeof e.composedPath === "function") {
    return e.composedPath().some(
      (node) => node instanceof Element && node.localName === localName,
    );
  }

  const target = e.target as Element | null;
  return !!target?.closest?.(localName);
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
