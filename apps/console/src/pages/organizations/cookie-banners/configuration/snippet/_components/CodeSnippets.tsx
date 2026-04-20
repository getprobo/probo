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
import { Button, Card, useToast } from "@probo/ui";
import { useState } from "react";
import { useParams } from "react-router";

export function CodeSnippets() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();

  const baseUrl = `${window.location.origin}/api/cookie-banner/v1`;

  const tabs = [
    {
      label: __("Script Tag"),
      code: `<script
  src="https://cdn.jsdelivr.net/npm/@probo/cookie-banner/dist/cookie-banner.iife.js"
  data-banner-id="${cookieBannerId}"
  data-base-url="${baseUrl}"
  data-position="bottom-left"
></script>`,
    },
    {
      label: __("ES Module"),
      code: `import { registerThemedBanner } from "@probo/cookie-banner/themed-banner";

registerThemedBanner();

// In your HTML or template:
// <probo-cookie-banner
//   banner-id="${cookieBannerId}"
//   base-url="${baseUrl}"
//   position="bottom-left"
// ></probo-cookie-banner>`,
    },
    {
      label: __("Headless"),
      code: `import { registerComponents } from "@probo/cookie-banner";

registerComponents();

// Build your own UI with headless components:
// <probo-cookie-banner-root banner-id="${cookieBannerId}" base-url="${baseUrl}">
//   <probo-banner>
//     <probo-accept-button><button>Accept all</button></probo-accept-button>
//     <probo-reject-button><button>Reject all</button></probo-reject-button>
//     <probo-customize-button><button>Customize</button></probo-customize-button>
//   </probo-banner>
//   <probo-settings-button position="bottom-left"></probo-settings-button>
// </probo-cookie-banner-root>`,
    },
  ];

  const [activeTab, setActiveTab] = useState(0);
  const activeCode = tabs[activeTab].code;

  const handleCopy = () => {
    void navigator.clipboard.writeText(activeCode);
    toast({
      title: __("Copied"),
      description: __("Code copied to clipboard"),
      variant: "success",
    });
  };

  return (
    <Card className="rounded-lg border">
      <div className="flex items-center justify-between border-b border-border-low px-1">
        <div className="flex">
          {tabs.map((tab, i) => (
            <button
              key={tab.label}
              type="button"
              onClick={() => setActiveTab(i)}
              className={`cursor-pointer px-3 py-2.5 text-sm font-light border-b-2 border-border-low -mb-px transition-colors ${
                i === activeTab
                  ? "border-border-mid text-foreground font-semibold"
                  : "border-transparent text-muted-foreground hover:text-foreground"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
        <Button variant="secondary" onClick={handleCopy}>
          {__("Copy")}
        </Button>
      </div>
      <pre className="overflow-x-auto p-4 text-sm font-mono bg-muted/30 rounded-b-lg text-invert bg-accent">
        <code>{activeCode}</code>
      </pre>
    </Card>
  );
}
