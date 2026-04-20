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
import { Card } from "@probo/ui";
import type { ReactNode } from "react";

import { ThemePreview } from "#/pages/organizations/cookie-banners/configuration/theme/_components/ThemePreview";

import { CodeSnippets } from "./_components/CodeSnippets";

export default function CookieBannerSnippetPage() {
  const { __ } = useTranslate();

  return (
    <div className="space-y-10">
      <Step
        number={1}
        title={__("Add the cookie banner snippet")}
        description={__("Include the Probo cookie banner on your website by adding one of the following snippets. The banner will automatically appear and collect visitor consent.")}
      >
        <CodeSnippets />
      </Step>

      <Step
        number={2}
        title={__("Customize the theme")}
        description={__("Adjust colors, fonts, and spacing to match your brand. The generated CSS snippet can be added to your website to override the default banner styles.")}
      >
        <ThemePreview />
      </Step>

      <Step
        number={3}
        title={__("Tag third-party elements with consent categories")}
        description={__("Mark scripts, iframes, and other third-party resources with a data-cookie-consent attribute so they only load after the visitor grants consent for the corresponding category. Replace src with data-src (or href with data-href) to prevent the browser from loading the resource before consent is given.")}
      >
        <div className="space-y-4">
          <Card className="border">
            <pre className="overflow-x-auto p-4 text-sm font-mono text-invert bg-accent rounded-lg">
              <code>
                {`<!-- Before: loads immediately -->
<script src="https://analytics.example.com/tracker.js"></script>

<!-- After: loads only when "analytics" consent is granted -->
<script
  type="text/plain"
  data-cookie-consent="analytics"
  data-src="https://analytics.example.com/tracker.js"
></script>`}
              </code>
            </pre>
          </Card>

          <p className="text-sm text-txt-secondary">
            {__("The same approach works for iframes, images, stylesheets, and other elements. See our documentation for the full list of supported elements and detailed integration guides.")}
          </p>
        </div>
      </Step>
    </div>
  );
}

interface StepProps {
  number: number;
  title: string;
  description: string;
  children: ReactNode;
}

function Step({ number, title, description, children }: StepProps) {
  return (
    <div className="flex gap-4">
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-accent text-sm font-semibold text-invert">
        {number}
      </div>
      <div className="min-w-0 flex-1 space-y-4">
        <div>
          <h3 className="text-lg font-medium">{title}</h3>
          <p className="mt-1 text-sm text-txt-secondary">{description}</p>
        </div>
        {children}
      </div>
    </div>
  );
}
