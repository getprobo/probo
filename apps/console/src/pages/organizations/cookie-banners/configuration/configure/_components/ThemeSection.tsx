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

import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useMemo, useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ThemeSection_cookieBanner$key } from "#/__generated__/core/ThemeSection_cookieBanner.graphql";

import { cookieBannerThemeSection } from "../../../variants";

type CSSVariable = {
  key: string;
  defaultValue: string;
  type: "color" | "text";
};

const CSS_VARIABLES: CSSVariable[] = [
  { key: "--probo-bg", defaultValue: "#ffffff", type: "color" },
  { key: "--probo-text", defaultValue: "#1a1a1a", type: "color" },
  { key: "--probo-text-secondary", defaultValue: "#555555", type: "color" },
  { key: "--probo-border", defaultValue: "#e0e0e0", type: "color" },
  { key: "--probo-accent", defaultValue: "#1a1a1a", type: "color" },
  { key: "--probo-accent-text", defaultValue: "#ffffff", type: "color" },
  { key: "--probo-radius", defaultValue: "12px", type: "text" },
  { key: "--probo-btn-radius", defaultValue: "8px", type: "text" },
  { key: "--probo-font-size", defaultValue: "14px", type: "text" },
  { key: "--probo-font-family", defaultValue: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, Helvetica, Arial, sans-serif", type: "text" },
  { key: "--probo-shadow", defaultValue: "0 4px 24px rgba(0, 0, 0, 0.12)", type: "text" },
];

function defaultValues(): Record<string, string> {
  const initial: Record<string, string> = {};
  for (const variable of CSS_VARIABLES) {
    initial[variable.key] = variable.defaultValue;
  }
  return initial;
}

function variableLabelKey(key: string): string {
  return key.replace("--probo-", "").replaceAll("-", "");
}

function buildCSSSnippet(values: Record<string, string>, defaultComment: string): string {
  const overrides = CSS_VARIABLES
    .filter(variable => values[variable.key] !== variable.defaultValue)
    .map(variable => `  ${variable.key}: ${values[variable.key]};`);

  if (overrides.length === 0) {
    return defaultComment;
  }

  return `probo-cookie-banner {\n${overrides.join("\n")}\n}`;
}

const themeSectionFragment = graphql`
  fragment ThemeSection_cookieBanner on CookieBanner {
    showBranding
  }
`;

interface ThemeSectionProps {
  cookieBannerKey: ThemeSection_cookieBanner$key;
}

export function ThemeSection({ cookieBannerKey }: ThemeSectionProps) {
  const cookieBanner = useFragment(themeSectionFragment, cookieBannerKey);
  const { t } = useTranslation("organizations/cookie-banners");
  const { toast } = useToast();
  const {
    root,
    intro,
    heading,
    fields,
    colors,
    color,
    colorRow,
    swatch,
    texts,
    preview,
    snippet,
    snippetBar,
  } = cookieBannerThemeSection();
  const [values, setValues] = useState<Record<string, string>>(defaultValues);

  const setValue = useCallback((key: string, value: string) => {
    setValues(previous => ({ ...previous, [key]: value }));
  }, []);

  const cssSnippet = useMemo(
    () => buildCSSSnippet(values, t("themeSection.defaultSnippet")),
    [values, t],
  );

  const previewStyle = useMemo(() => {
    const style: Record<string, string> = {};
    for (const variable of CSS_VARIABLES) {
      style[variable.key] = values[variable.key];
    }
    return style;
  }, [values]);

  function handleCopyCSS() {
    void navigator.clipboard.writeText(cssSnippet).then(
      () => {
        toast({
          title: t("themeSection.messages.copiedTitle"),
          description: t("themeSection.messages.copied"),
          variant: "success",
        });
      },
      () => {
        toast({
          title: t("themeSection.errors.title"),
          description: t("themeSection.errors.copy"),
          variant: "error",
        });
      },
    );
  }

  const colorVariables = CSS_VARIABLES.filter(variable => variable.type === "color");
  const textVariables = CSS_VARIABLES.filter(variable => variable.type === "text");

  return (
    <section className={root()}>
      <div className={intro()}>
        <div className={heading()}>
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("themeSection.title")}
          </Heading>
          <Text size={2} color="faint">
            {t("themeSection.description")}
          </Text>
        </div>
        <Button size={2} variant="soft" color="neutral" onClick={() => setValues(defaultValues())}>
          {t("themeSection.actions.reset")}
        </Button>
      </div>
      <Card size={2} variant="soft">
        <div className={fields()}>
          <div className={colors()}>
            {colorVariables.map(variable => (
              <div key={variable.key} className={color()}>
                <Text size={2} weight="medium" highContrast>
                  {t(`themeSection.variables.${variableLabelKey(variable.key)}`)}
                </Text>
                <div className={colorRow()}>
                  <input
                    type="color"
                    value={values[variable.key]}
                    className={swatch()}
                    aria-label={t(`themeSection.variables.${variableLabelKey(variable.key)}`)}
                    onChange={event => setValue(variable.key, event.target.value)}
                  />
                  <TextField
                    size={2}
                    value={values[variable.key]}
                    aria-label={t(`themeSection.variables.${variableLabelKey(variable.key)}`)}
                    onValueChange={value => setValue(variable.key, value)}
                  />
                </div>
              </div>
            ))}
          </div>
          <div className={texts()}>
            {textVariables.map(variable => (
              <Field
                key={variable.key}
                label={t(`themeSection.variables.${variableLabelKey(variable.key)}`)}
              >
                <TextField
                  size={2}
                  value={values[variable.key]}
                  onValueChange={value => setValue(variable.key, value)}
                />
              </Field>
            ))}
          </div>
        </div>
      </Card>
      <Card size={2} variant="soft" padding="none">
        <div className={preview()} style={previewStyle}>
          <BannerPreview showBranding={cookieBanner.showBranding} />
        </div>
      </Card>
      <div className={snippetBar()}>
        <Button size={2} variant="soft" color="neutral" onClick={handleCopyCSS}>
          {t("themeSection.actions.copy")}
        </Button>
      </div>
      <Card size={2} variant="soft">
        <pre className={snippet()}>
          <code>{cssSnippet}</code>
        </pre>
      </Card>
    </section>
  );
}

function BannerPreview({ showBranding }: { showBranding: boolean }) {
  const { t } = useTranslation("organizations/cookie-banners");

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
        padding: "24px 24px 12px 24px",
      }}
    >
      <p
        style={{
          fontSize: "calc(var(--probo-font-size, 14px) + 2px)",
          fontWeight: 600,
          margin: "0 0 8px",
        }}
      >
        {t("themeSection.banner.title")}
      </p>
      <p
        style={{
          color: "var(--probo-text-secondary, #555555)",
          margin: "0 0 20px",
        }}
      >
        <Trans
          ns="organizations/cookie-banners"
          i18nKey="themeSection.banner.description"
          components={{
            policy: (
              <a
                href="#"
                onClick={event => event.preventDefault()}
                style={{
                  color: "var(--probo-accent, #1a1a1a)",
                  textDecoration: "underline",
                }}
              />
            ),
          }}
        />
      </p>
      <div style={{ display: "flex", gap: 8, flexWrap: "wrap", paddingBottom: "12px" }}>
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
          {t("themeSection.banner.acceptAll")}
        </button>
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
          {t("themeSection.banner.rejectAll")}
        </button>
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
          {t("themeSection.banner.customize")}
        </button>
      </div>
      {showBranding && (
        <div
          style={{
            textAlign: "center",
            fontSize: "calc(var(--probo-font-size, 14px) - 2px)",
            fontWeight: 400,
            color: "var(--probo-text-secondary, #555555)",
          }}
        >
          {t("themeSection.banner.privacyBy")}
          {" "}
          Probo
        </div>
      )}
    </div>
  );
}
