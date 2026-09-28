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

import { ArrowUpRightIcon } from "@phosphor-icons/react";
import { externalLinkProps } from "@probo/helpers";
import { Anchor } from "@probo/ui/src/v2/Link/Anchor";
import type { PropsWithChildren } from "react";

type ConnectorCredentialPageLinkProps = PropsWithChildren<{ url: string | null }>;

// A link to the vendor page where the customer creates the credential a
// connect dialog asks for, or nothing until that page is known.
export function ConnectorCredentialPageLink({
  url,
  children,
}: ConnectorCredentialPageLinkProps) {
  if (!url) {
    return null;
  }

  const link = externalLinkProps(url);
  if (!link.href) {
    return null;
  }

  return (
    <Anchor
      {...link}
      size={2}
      iconEnd={<ArrowUpRightIcon aria-hidden />}
    >
      {children}
    </Anchor>
  );
}
