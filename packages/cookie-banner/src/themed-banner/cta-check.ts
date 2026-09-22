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

const MIN_CONTRAST = 5;
const PRIMARY_CTA_BUTTONS =
  "probo-accept-button button, probo-reject-button button, probo-acknowledge-button button";
const warned = new WeakSet<object>();

interface RGB {
  r: number;
  g: number;
  b: number;
}

interface RGBA extends RGB {
  a: number;
}

export interface CtaComputedStyle {
  color: string;
  backgroundColor: string;
  fontFamily: string;
  fontSize: string;
  fontWeight: string;
  fontStyle: string;
  textTransform: string;
  textDecorationLine: string;
  letterSpacing: string;
}

export interface NamedCtaStyle {
  name: string;
  style: CtaComputedStyle;
}

const TREATMENT: {
  key: keyof CtaComputedStyle;
  label: string;
  normalize: (value: string) => string;
}[] = [
  { key: "fontFamily", label: "font-family", normalize: normalizeFamily },
  { key: "fontSize", label: "font-size", normalize: lower },
  { key: "fontWeight", label: "font-weight", normalize: normalizeWeight },
  { key: "fontStyle", label: "font-style", normalize: lower },
  { key: "textTransform", label: "text-transform", normalize: lower },
  { key: "textDecorationLine", label: "text-decoration", normalize: normalizeDecoration },
  { key: "letterSpacing", label: "letter-spacing", normalize: normalizeSpacing },
];

export function parseCssColor(value: string): RGBA | undefined {
  const raw = value.trim().toLowerCase();
  if (!raw || raw === "currentcolor") {
    return undefined;
  }
  if (raw === "transparent") {
    return { r: 0, g: 0, b: 0, a: 0 };
  }
  if (raw.startsWith("#")) {
    return parseHex(raw);
  }
  const rgb = /^rgba?\(\s*([^\s,/)]+)(?:\s*,\s*|\s+)([^\s,/)]+)(?:\s*,\s*|\s+)([^\s,/)]+)(?:\s*[,/]\s*([^\s)]+))?\s*\)$/.exec(
    raw,
  );
  if (!rgb) {
    return undefined;
  }
  const r = parseChannel(rgb[1]);
  const g = parseChannel(rgb[2]);
  const b = parseChannel(rgb[3]);
  if (r === undefined || g === undefined || b === undefined) {
    return undefined;
  }
  return { r, g, b, a: parseAlpha(rgb[4]) };
}

export function contrastRatio(fg: RGB, bg: RGB): number {
  const lighter = Math.max(relativeLuminance(fg), relativeLuminance(bg));
  const darker = Math.min(relativeLuminance(fg), relativeLuminance(bg));
  return (lighter + 0.05) / (darker + 0.05);
}

export function firstLayerCtaIssues(ctas: NamedCtaStyle[]): string[] {
  const issues: string[] = [];
  if (ctas.length >= 2) {
    const [primary, secondary] = ctas;
    const mismatched = TREATMENT.filter(
      (field) =>
        field.normalize(String(primary.style[field.key] ?? "")) !==
        field.normalize(String(secondary.style[field.key] ?? "")),
    ).map((field) => field.label);
    if (mismatched.length > 0) {
      issues.push(`text treatment differs (${mismatched.join(", ")})`);
    }
  }

  for (const cta of ctas) {
    const ratio = contrastOf(cta.style);
    if (ratio === undefined) {
      issues.push(`${cta.name} contrast cannot be verified`);
      continue;
    }
    if (ratio < MIN_CONTRAST) {
      issues.push(`${cta.name} contrast ${formatRatio(ratio)} (minimum ${MIN_CONTRAST}:1)`);
    }
  }
  return issues;
}

export function formatFirstLayerCtaWarning(issues: string[]): string | undefined {
  if (issues.length === 0) {
    return undefined;
  }
  return (
    "First-layer CTA checks failed: " +
    issues.join("; ") +
    ". Custom CSS variables or a style snippet can cause this."
  );
}

export function scheduleFirstLayerCtaCheck(root: ParentNode): void {
  const run = () => warnFirstLayerCtas(root);
  if (typeof requestAnimationFrame === "function") {
    requestAnimationFrame(run);
    return;
  }
  run();
}

export function warnFirstLayerCtas(root: ParentNode): void {
  if (warned.has(root)) {
    return;
  }
  warned.add(root);

  const banner = root.querySelector("probo-banner");
  if (!banner) {
    return;
  }

  const ctas: NamedCtaStyle[] = [];
  for (const button of banner.querySelectorAll(PRIMARY_CTA_BUTTONS)) {
    const style = readCtaComputedStyle(button);
    if (!style) {
      continue;
    }
    ctas.push({ name: ctaName(button), style });
  }
  if (ctas.length === 0) {
    return;
  }

  const message = formatFirstLayerCtaWarning(firstLayerCtaIssues(ctas));
  if (message) {
    console.warn(`[probo] ${message}`);
  }
}

function ctaName(button: Element): string {
  if (button.closest("probo-reject-button")) {
    return "secondary CTA";
  }
  return "primary CTA";
}

function readCtaComputedStyle(el: Element): CtaComputedStyle | undefined {
  if (typeof getComputedStyle !== "function") {
    return undefined;
  }
  const style = getComputedStyle(el);
  const background = resolvedBackground(el);
  return {
    color: style.color,
    backgroundColor: cssRgb(background),
    fontFamily: style.fontFamily,
    fontSize: style.fontSize,
    fontWeight: String(style.fontWeight),
    fontStyle: style.fontStyle,
    textTransform: style.textTransform,
    textDecorationLine: style.textDecorationLine || decorationLine(style.textDecoration),
    letterSpacing: style.letterSpacing,
  };
}

function resolvedBackground(el: Element): RGB {
  const layers: RGBA[] = [];
  let node: Element | null = el;
  while (node) {
    const color = parseCssColor(getComputedStyle(node).backgroundColor);
    if (color && color.a > 0) {
      layers.push(color);
      if (color.a >= 0.99) {
        break;
      }
    }
    node = nextAncestor(node);
  }
  let result: RGB = { r: 255, g: 255, b: 255 };
  for (let i = layers.length - 1; i >= 0; i--) {
    result = composite(layers[i], result);
  }
  return result;
}

function nextAncestor(node: Element): Element | null {
  if (node.parentElement) {
    return node.parentElement;
  }
  const root = node.getRootNode();
  if (root instanceof ShadowRoot) {
    return root.host;
  }
  return null;
}

function contrastOf(style: CtaComputedStyle): number | undefined {
  const fg = parseCssColor(style.color);
  const bg = parseCssColor(style.backgroundColor);
  if (!fg || !bg) {
    return undefined;
  }
  const opaqueBg = composite(bg, { r: 255, g: 255, b: 255 });
  return contrastRatio(composite(fg, opaqueBg), opaqueBg);
}

function composite(src: RGBA, dst: RGB): RGB {
  const a = clamp(src.a, 0, 1);
  return {
    r: src.r * a + dst.r * (1 - a),
    g: src.g * a + dst.g * (1 - a),
    b: src.b * a + dst.b * (1 - a),
  };
}

function relativeLuminance(color: RGB): number {
  return 0.2126 * srgb(color.r) + 0.7152 * srgb(color.g) + 0.0722 * srgb(color.b);
}

function srgb(channel: number): number {
  const c = clamp(channel, 0, 255) / 255;
  return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
}

function parseHex(value: string): RGBA | undefined {
  const hex = value.slice(1);
  if (hex.length === 3 || hex.length === 4) {
    const r = Number.parseInt(hex[0] + hex[0], 16);
    const g = Number.parseInt(hex[1] + hex[1], 16);
    const b = Number.parseInt(hex[2] + hex[2], 16);
    const a = hex.length === 4 ? Number.parseInt(hex[3] + hex[3], 16) / 255 : 1;
    if ([r, g, b, a].some((n) => Number.isNaN(n))) {
      return undefined;
    }
    return { r, g, b, a };
  }
  if (hex.length === 6 || hex.length === 8) {
    const r = Number.parseInt(hex.slice(0, 2), 16);
    const g = Number.parseInt(hex.slice(2, 4), 16);
    const b = Number.parseInt(hex.slice(4, 6), 16);
    const a = hex.length === 8 ? Number.parseInt(hex.slice(6, 8), 16) / 255 : 1;
    if ([r, g, b, a].some((n) => Number.isNaN(n))) {
      return undefined;
    }
    return { r, g, b, a };
  }
  return undefined;
}

function parseChannel(raw: string): number | undefined {
  const n = Number.parseFloat(raw);
  if (!Number.isFinite(n)) {
    return undefined;
  }
  if (raw.includes("%")) {
    return clamp(n / 100, 0, 1) * 255;
  }
  return clamp(n, 0, 255);
}

function parseAlpha(raw: string | undefined): number {
  if (raw === undefined || raw === "") {
    return 1;
  }
  const n = Number.parseFloat(raw);
  if (!Number.isFinite(n)) {
    return 1;
  }
  if (raw.includes("%")) {
    return clamp(n / 100, 0, 1);
  }
  return clamp(n, 0, 1);
}

function formatRatio(ratio: number): string {
  return `${(Math.floor(ratio * 10) / 10).toString()}:1`;
}

function cssRgb(color: RGB): string {
  return `rgb(${Math.round(color.r)}, ${Math.round(color.g)}, ${Math.round(color.b)})`;
}

function decorationLine(value: string): string {
  const token = value.trim().split(/\s+/)[0]?.toLowerCase();
  if (!token || token === "none" || token.startsWith("rgb") || token.startsWith("#")) {
    return "none";
  }
  return token;
}

function normalizeFamily(value: string): string {
  return value
    .replace(/["']/g, "")
    .replace(/\s*,\s*/g, ",")
    .replace(/\s+/g, " ")
    .trim()
    .toLowerCase();
}

function normalizeWeight(value: string): string {
  const v = value.trim().toLowerCase();
  if (v === "normal") {
    return "400";
  }
  if (v === "bold") {
    return "700";
  }
  return v;
}

function normalizeDecoration(value: string): string {
  const v = value.trim().toLowerCase();
  if (!v || v === "none") {
    return "none";
  }
  return decorationLine(v);
}

function normalizeSpacing(value: string): string {
  const v = value.trim().toLowerCase();
  if (v === "normal" || v === "0" || v === "0px") {
    return "0";
  }
  return v;
}

function lower(value: string): string {
  return value.trim().toLowerCase();
}

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n));
}
