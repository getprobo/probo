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

import { useTranslate } from "@probo/i18n";
import { Button, IconArrowLink } from "@probo/ui";

type Props = {
  url?: string | null;
  // "link" (default) is a quiet inline text link for the provider card;
  // "button" is a secondary button that sits next to Cancel/Connect in a
  // connect dialog footer. Both open the docs page in a new tab.
  variant?: "link" | "button";
};

// A "Documentation" link to a connector's probo.com docs page, or nothing when
// the provider has no documentation URL. Shared by the provider card and the
// connect dialogs so the URL/label/target markup lives in one place.
export function ConnectorDocumentationLink({ url, variant = "link" }: Props) {
  const { __ } = useTranslate();

  if (!url) {
    return null;
  }

  if (variant === "button") {
    // asChild styles the anchor as a secondary button (matching Cancel); the
    // button base supplies the flex + gap that spaces the label and icon.
    return (
      <Button variant="secondary" asChild>
        <a href={url} target="_blank" rel="noopener noreferrer">
          {__("Documentation")}
          <IconArrowLink size={16} />
        </a>
      </Button>
    );
  }

  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="inline-flex items-center gap-1 text-xs text-txt-tertiary underline hover:no-underline"
    >
      {__("Documentation")}
      <IconArrowLink size={12} />
    </a>
  );
}
