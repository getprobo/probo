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
import { useParams } from "react-router";

export function CodeSnippets() {
  const { __ } = useTranslate();
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
    void navigator.clipboard.writeText(code);
    toast({
      title: __("Copied"),
      description: __("Code copied to clipboard"),
      variant: "success",
    });
  };

  return (
    <div className="space-y-3">
      <Card className="rounded-lg border">
        <div className="flex items-center justify-end border-b border-border-low px-1 py-1">
          <Button variant="secondary" onClick={handleCopy}>
            {__("Copy")}
          </Button>
        </div>
        <pre className="overflow-x-auto p-4 text-sm font-mono rounded-b-lg text-invert bg-accent">
          <code>{code}</code>
        </pre>
      </Card>

      <p className="text-sm text-txt-secondary">
        {__("Looking for ES module or headless integration?")}
        {" "}
        <a
          href="https://docs.getprobo.com/guides/cookie-banner"
          target="_blank"
          rel="noopener noreferrer"
          className="text-txt-primary underline hover:no-underline"
        >
          {__("See our documentation for detailed integration guides.")}
        </a>
      </p>
    </div>
  );
}
