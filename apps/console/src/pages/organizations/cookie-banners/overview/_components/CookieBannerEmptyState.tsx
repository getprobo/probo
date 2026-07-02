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

import { CookieIcon } from "@phosphor-icons/react";
import { useTranslate } from "@probo/i18n";
import type { ReactNode } from "react";

interface CookieBannerEmptyStateProps {
  children?: ReactNode;
}

export function CookieBannerEmptyState({ children }: CookieBannerEmptyStateProps) {
  const { __ } = useTranslate();

  const steps = [
    {
      step: "1",
      title: __("Create a banner"),
      description: __("Set up your cookie consent banner with a name, origin URL, and privacy policy link."),
    },
    {
      step: "2",
      title: __("Configure categories"),
      description: __("Organize your cookies into categories like Analytics, Advertising, and Functional."),
    },
    {
      step: "3",
      title: __("Install the SDK"),
      description: __("Add a single script tag or import the ES module to start collecting consent."),
    },
  ];

  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <CookieIcon size={48} weight="duotone" className="mb-2 text-muted-foreground" />
      <h2 className="text-xl font-semibold mb-2">{__("No cookie banners yet")}</h2>
      <p className="text-muted-foreground mb-8 max-w-md">
        {__("Create your first cookie consent banner to start collecting GDPR-compliant consent from your website visitors.")}
      </p>

      <div className="grid gap-6 sm:grid-cols-3 mb-8 w-full max-w-2xl">
        {steps.map(s => (
          <div key={s.step} className="rounded-lg border border-border-mid p-4 text-left">
            <div className="mb-2 flex size-8 items-center justify-center rounded-full bg-border-solid text-primary-foreground text-sm font-semibold">
              {s.step}
            </div>
            <h3 className="font-medium mb-1">{s.title}</h3>
            <p className="text-sm text-muted-foreground">{s.description}</p>
          </div>
        ))}
      </div>

      {children}
    </div>
  );
}
