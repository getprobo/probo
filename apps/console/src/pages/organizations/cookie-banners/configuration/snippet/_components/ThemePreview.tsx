// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { Button, Card, Field, Input, useToast } from "@probo/ui";
import { useCallback, useMemo, useState } from "react";

type CSSVariable = {
  key: string;
  label: string;
  defaultValue: string;
  type: "color" | "text";
};

const CSS_VARIABLES: CSSVariable[] = [
  { key: "--probo-bg", label: "Background", defaultValue: "#ffffff", type: "color" },
  { key: "--probo-text", label: "Text", defaultValue: "#1a1a1a", type: "color" },
  { key: "--probo-text-secondary", label: "Text Secondary", defaultValue: "#555555", type: "color" },
  { key: "--probo-border", label: "Border", defaultValue: "#e0e0e0", type: "color" },
  { key: "--probo-accent", label: "Accent", defaultValue: "#1a1a1a", type: "color" },
  { key: "--probo-accent-text", label: "Accent Text", defaultValue: "#ffffff", type: "color" },
  { key: "--probo-radius", label: "Border Radius", defaultValue: "12px", type: "text" },
  { key: "--probo-btn-radius", label: "Button Radius", defaultValue: "8px", type: "text" },
  { key: "--probo-font-size", label: "Font Size", defaultValue: "14px", type: "text" },
  { key: "--probo-font-family", label: "Font Family", defaultValue: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, Helvetica, Arial, sans-serif", type: "text" },
  { key: "--probo-shadow", label: "Shadow", defaultValue: "0 4px 24px rgba(0, 0, 0, 0.12)", type: "text" },
];

function buildCSSSnippet(values: Record<string, string>): string {
  const overrides = CSS_VARIABLES
    .filter(v => values[v.key] !== v.defaultValue)
    .map(v => `  ${v.key}: ${values[v.key]};`);

  if (overrides.length === 0) {
    return "/* Using default theme — no overrides needed */";
  }

  return `probo-cookie-banner {\n${overrides.join("\n")}\n}`;
}

export function ThemePreview() {
  const { __ } = useTranslate();
  const { toast } = useToast();

  const [values, setValues] = useState<Record<string, string>>(() => {
    const initial: Record<string, string> = {};
    for (const v of CSS_VARIABLES) {
      initial[v.key] = v.defaultValue;
    }
    return initial;
  });

  const setValue = useCallback((key: string, value: string) => {
    setValues(prev => ({ ...prev, [key]: value }));
  }, []);

  const handleReset = useCallback(() => {
    const initial: Record<string, string> = {};
    for (const v of CSS_VARIABLES) {
      initial[v.key] = v.defaultValue;
    }
    setValues(initial);
  }, []);

  const cssSnippet = useMemo(() => buildCSSSnippet(values), [values]);

  const handleCopyCSS = () => {
    void navigator.clipboard.writeText(cssSnippet);
    toast({
      title: __("Copied"),
      description: __("CSS snippet copied to clipboard"),
      variant: "success",
    });
  };

  const previewStyle = useMemo(() => {
    const style: Record<string, string> = {};
    for (const v of CSS_VARIABLES) {
      style[v.key] = values[v.key];
    }
    return style;
  }, [values]);

  const colorVariables = CSS_VARIABLES.filter(v => v.type === "color");
  const textVariables = CSS_VARIABLES.filter(v => v.type === "text");

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium text-lg">{__("Theme")}</h3>
        <Button variant="tertiary" onClick={handleReset}>
          {__("Reset to defaults")}
        </Button>
      </div>

      <Card className="border p-4">
        <div className="space-y-4">
          <div className="grid grid-cols-3 gap-4">
            {colorVariables.map(v => (
              <div key={v.key} className="flex flex-col gap-1.5">
                <label className="text-sm font-medium text-txt-primary">
                  {v.label}
                </label>
                <div className="flex items-center gap-2">
                  <input
                    type="color"
                    value={values[v.key]}
                    onChange={e => setValue(v.key, e.target.value)}
                    className="h-8 w-10 shrink-0 cursor-pointer rounded border border-border-mid bg-transparent p-0.5"
                  />
                  <Input
                    value={values[v.key]}
                    onChange={e => setValue(v.key, e.target.value)}
                  />
                </div>
              </div>
            ))}
          </div>

          <div className="grid grid-cols-2 gap-4">
            {textVariables.map(v => (
              <Field key={v.key} label={v.label}>
                <Input
                  value={values[v.key]}
                  onChange={e => setValue(v.key, e.target.value)}
                />
              </Field>
            ))}
          </div>
        </div>
      </Card>

      <h3 className="font-medium text-lg">{__("Preview")}</h3>

      <Card className="border overflow-hidden">
        <div
          className="relative flex items-end justify-center bg-[repeating-conic-gradient(#e5e7eb_0%_25%,transparent_0%_50%)] bg-size-[20px_20px] p-8"
          style={{ minHeight: 280, ...previewStyle }}
        >
          <BannerPreview />
        </div>
      </Card>

      <div className="flex items-center justify-between">
        <h3 className="font-medium text-lg">{__("CSS Snippet")}</h3>
        <Button variant="secondary" onClick={handleCopyCSS}>
          {__("Copy")}
        </Button>
      </div>

      <Card className="border">
        <pre className="overflow-x-auto p-4 text-sm font-mono text-invert bg-accent rounded-lg">
          <code>{cssSnippet}</code>
        </pre>
      </Card>
    </div>
  );
}

function BannerPreview() {
  return (
    <div
      style={{
        background: "var(--probo-bg, #ffffff)",
        color: "var(--probo-text, #1a1a1a)",
        borderRadius: "var(--probo-radius, 12px)",
        boxShadow: "var(--probo-shadow, 0 4px 24px rgba(0, 0, 0, 0.12))",
        fontFamily: "var(--probo-font-family, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif)",
        fontSize: "var(--probo-font-size, 14px)",
        lineHeight: 1.5,
        maxWidth: 450,
        width: "100%",
        padding: 24,
      }}
    >
      <p
        style={{
          fontSize: "calc(var(--probo-font-size, 14px) + 2px)",
          fontWeight: 600,
          margin: "0 0 8px",
        }}
      >
        Cookie Preferences
      </p>
      <p
        style={{
          color: "var(--probo-text-secondary, #555555)",
          margin: "0 0 20px",
        }}
      >
        We use cookies to improve your experience and analyze site traffic.
        {" "}
        <a
          href="#"
          onClick={e => e.preventDefault()}
          style={{
            color: "var(--probo-accent, #1a1a1a)",
            textDecoration: "underline",
          }}
        >
          Privacy Policy
        </a>
      </p>
      <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
        <span>
          <button
            type="button"
            style={{
              padding: "8px 10px",
              borderRadius: "var(--probo-btn-radius, 8px)",
              border: "1px solid var(--probo-accent, #1a1a1a)",
              background: "var(--probo-accent, #1a1a1a)",
              color: "var(--probo-accent-text, #ffffff)",
              fontFamily: "inherit",
              fontSize: "var(--probo-font-size, 14px)",
              fontWeight: 500,
              lineHeight: "normal",
              cursor: "pointer",
              whiteSpace: "nowrap",
            }}
          >
            Accept all
          </button>
        </span>
        <span>
          <button
            type="button"
            style={{
              padding: "8px 10px",
              borderRadius: "var(--probo-btn-radius, 8px)",
              border: "1px solid var(--probo-border, #e0e0e0)",
              background: "color-mix(in srgb, var(--probo-text, #1a1a1a) 8%, var(--probo-bg, #ffffff))",
              color: "var(--probo-text, #1a1a1a)",
              fontFamily: "inherit",
              fontSize: "var(--probo-font-size, 14px)",
              fontWeight: 500,
              lineHeight: "normal",
              cursor: "pointer",
              whiteSpace: "nowrap",
            }}
          >
            Reject all
          </button>
        </span>
        <span>
          <button
            type="button"
            style={{
              padding: "8px 10px",
              borderRadius: "var(--probo-btn-radius, 8px)",
              border: "none",
              background: "transparent",
              color: "var(--probo-accent, #1a1a1a)",
              fontFamily: "inherit",
              fontSize: "var(--probo-font-size, 14px)",
              fontWeight: 500,
              lineHeight: "normal",
              cursor: "pointer",
              whiteSpace: "nowrap",
              textDecoration: "underline",
            }}
          >
            Customize
          </button>
        </span>
      </div>
    </div>
  );
}
