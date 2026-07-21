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

import { Button, Card, useToast } from "@probo/ui";
import { Trans, useTranslation } from "react-i18next";
import { useParams } from "react-router";

export function CodeSnippets() {
  const { t } = useTranslation("organizations/cookie-banners");
  const { toast } = useToast();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();

  const baseUrl = `${window.location.origin}/api/cookie-banner/v1`;

  const code = `<script
  src="https://cdn.jsdelivr.net/npm/@probo/cookie-banner/dist/cookie-banner.iife.js"
  data-banner-id="${cookieBannerId}"
  data-base-url="${baseUrl}"
  data-position="bottom-left"
></script>`;

  const handleCopy = () => {
    navigator.clipboard.writeText(code).then(
      () => {
        toast({
          title: t("codeSnippets.messages.copiedTitle"),
          description: t("codeSnippets.messages.copied"),
          variant: "success",
        });
      },
      () => {
        toast({
          title: t("codeSnippets.errors.title"),
          description: t("codeSnippets.errors.copy"),
          variant: "error",
        });
      },
    );
  };

  return (
    <div className="space-y-3">
      <h3 className="font-medium">{t("codeSnippets.title")}</h3>
      <Card className="rounded-lg border">
        <div className="flex items-center justify-end border-b border-border-low px-1 py-1">
          <Button variant="secondary" onClick={handleCopy}>
            {t("codeSnippets.actions.copy")}
          </Button>
        </div>
        <pre className="overflow-x-auto p-4 text-sm font-mono rounded-b-lg text-invert bg-accent">
          <code>{code}</code>
        </pre>
      </Card>

      <p className="text-sm text-txt-secondary">
        <Trans
          ns="organizations/cookie-banners"
          i18nKey="codeSnippets.documentation"
          components={{
            link: (
              <a
                href="https://www.probo.com/docs/product/cookie-banner/javascript-sdk"
                target="_blank"
                rel="noopener noreferrer"
                className="text-txt-primary underline hover:no-underline"
              />
            ),
          }}
        />
      </p>
    </div>
  );
}
