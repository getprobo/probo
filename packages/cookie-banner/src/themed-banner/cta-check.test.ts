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

import { describe, expect, it } from "vitest";

import {
  contrastRatio,
  firstLayerCtaIssues,
  formatFirstLayerCtaWarning,
  parseCssColor,
  type CtaComputedStyle,
  type NamedCtaStyle,
} from "./cta-check";

const treatment = {
  fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif',
  fontSize: "14px",
  fontWeight: "500",
  fontStyle: "normal",
  textTransform: "none",
  textDecorationLine: "none",
  letterSpacing: "normal",
};

const defaultPrimary: CtaComputedStyle = {
  ...treatment,
  color: "rgb(255, 255, 255)",
  backgroundColor: "rgb(26, 26, 26)",
};

const defaultSecondary: CtaComputedStyle = {
  ...treatment,
  color: "rgb(26, 26, 26)",
  backgroundColor: "rgb(237, 237, 237)",
};

function pair(primary = defaultPrimary, secondary = defaultSecondary): NamedCtaStyle[] {
  return [
    { name: "primary CTA", style: primary },
    { name: "secondary CTA", style: secondary },
  ];
}

describe("parseCssColor", () => {
  it("parses hex, rgb, and transparent", () => {
    expect(parseCssColor("#1a1a1a")).toEqual({ r: 26, g: 26, b: 26, a: 1 });
    expect(parseCssColor("#fff")).toEqual({ r: 255, g: 255, b: 255, a: 1 });
    expect(parseCssColor("rgb(26, 26, 26)")).toEqual({ r: 26, g: 26, b: 26, a: 1 });
    expect(parseCssColor("rgb(26 26 26 / 80%)")).toEqual({ r: 26, g: 26, b: 26, a: 0.8 });
    expect(parseCssColor("transparent")).toEqual({ r: 0, g: 0, b: 0, a: 0 });
    expect(parseCssColor("color-mix(in srgb, black 8%, white)")).toBeUndefined();
  });
});

describe("contrastRatio", () => {
  it("is 21:1 for black on white", () => {
    expect(contrastRatio({ r: 0, g: 0, b: 0 }, { r: 255, g: 255, b: 255 })).toBe(21);
  });
});

describe("firstLayerCtaIssues", () => {
  it("accepts the default themed primary / secondary styles", () => {
    expect(firstLayerCtaIssues(pair())).toEqual([]);
  });

  it("allows different fill colors when typography matches", () => {
    expect(
      firstLayerCtaIssues(
        pair(defaultPrimary, {
          ...defaultSecondary,
          color: "rgb(0, 0, 0)",
          backgroundColor: "rgb(255, 255, 255)",
        }),
      ),
    ).toEqual([]);
  });

  it("flags mismatched text treatment", () => {
    expect(
      firstLayerCtaIssues(
        pair(defaultPrimary, {
          ...defaultSecondary,
          fontWeight: "700",
          fontSize: "16px",
        }),
      ),
    ).toEqual(["text treatment differs (font-size, font-weight)"]);
  });

  it("treats bold and 700 as the same weight", () => {
    expect(
      firstLayerCtaIssues(
        pair({ ...defaultPrimary, fontWeight: "bold" }, { ...defaultSecondary, fontWeight: "700" }),
      ),
    ).toEqual([]);
  });

  it("flags unverifiable wide-gamut colors instead of skipping them", () => {
    expect(
      firstLayerCtaIssues(
        pair({ ...defaultPrimary, color: "oklch(0.5 0.1 20)" }, defaultSecondary),
      ),
    ).toEqual(["primary CTA contrast cannot be verified"]);
  });

  it("does not round a failing contrast ratio up to 5:1", () => {
    const issues = firstLayerCtaIssues([
      {
        name: "primary CTA",
        style: {
          ...defaultPrimary,
          color: "rgb(112, 112, 112)",
          backgroundColor: "rgb(255, 255, 255)",
        },
      },
    ]);
    expect(issues).toHaveLength(1);
    expect(issues[0]).toMatch(/^primary CTA contrast 4\.\d:1 \(minimum 5:1\)$/);
    expect(issues[0]).not.toContain(" 5:1 (minimum");
  });

  it("flags contrast below 5:1 on either primary CTA", () => {
    const low = firstLayerCtaIssues(
      pair(
        { ...defaultPrimary, color: "rgb(255, 255, 255)", backgroundColor: "rgb(204, 204, 204)" },
        defaultSecondary,
      ),
    );
    expect(low).toHaveLength(1);
    expect(low[0]).toMatch(/^primary CTA contrast \d+(\.\d+)?:1 \(minimum 5:1\)$/);

    const both = firstLayerCtaIssues(
      pair(
        { ...defaultPrimary, color: "rgb(255, 255, 255)", backgroundColor: "rgb(204, 204, 204)" },
        { ...defaultSecondary, color: "rgb(187, 187, 187)", backgroundColor: "rgb(255, 255, 255)" },
      ),
    );
    expect(both).toHaveLength(2);
    expect(both[0]).toMatch(/^primary CTA contrast /);
    expect(both[1]).toMatch(/^secondary CTA contrast /);
  });

  it("checks contrast on a single notice CTA without text-treatment pairing", () => {
    expect(firstLayerCtaIssues([{ name: "primary CTA", style: defaultPrimary }])).toEqual([]);
    const low = firstLayerCtaIssues([
      {
        name: "primary CTA",
        style: { ...defaultPrimary, color: "rgb(255, 255, 255)", backgroundColor: "rgb(204, 204, 204)" },
      },
    ]);
    expect(low).toEqual(expect.arrayContaining([expect.stringMatching(/^primary CTA contrast /)]));
    expect(low.join(" ")).not.toContain("text treatment");
  });
});

describe("formatFirstLayerCtaWarning", () => {
  it("returns nothing when the CTAs pass", () => {
    expect(formatFirstLayerCtaWarning([])).toBeUndefined();
  });

  it("names style snippets as the likely cause", () => {
    expect(formatFirstLayerCtaWarning(["text treatment differs (font-weight)"])).toBe(
      "First-layer CTA checks failed: text treatment differs (font-weight). Custom CSS variables or a style snippet can cause this.",
    );
  });
});
