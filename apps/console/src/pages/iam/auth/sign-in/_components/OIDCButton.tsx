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
import { Button, Google, Microsoft } from "@probo/ui";
import type { ComponentProps } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { OIDCButtonFragment$key } from "#/__generated__/iam/OIDCButtonFragment.graphql";
import { useSafeContinueUrl } from "#/hooks/useSafeContinueUrl";

const fragment = graphql`
  fragment OIDCButtonFragment on OIDCProviderInfo {
    name
    loginURL
  }
`;

const providerIcons: Record<
  string,
  (props: ComponentProps<"svg">) => React.ReactNode
> = {
  google: Google,
  microsoft: Microsoft,
};

export function OIDCButton({
  providerRef,
}: {
  providerRef: OIDCButtonFragment$key;
}) {
  const { __ } = useTranslate();
  const safeContinueUrl = useSafeContinueUrl();
  const provider = useFragment(fragment, providerRef);
  const Icon = providerIcons[provider.name];

  return (
    <Button
      variant="secondary"
      className="w-full h-10"
      onClick={() => {
        window.location.href
          = provider.loginURL
            + "?continue="
            + encodeURIComponent(safeContinueUrl.toString());
      }}
    >
      <span className="flex items-center gap-2">
        {Icon && <Icon width={18} height={18} />}
        {__(`Sign in with ${provider.name.charAt(0).toUpperCase() + provider.name.slice(1)}`)}
      </span>
    </Button>
  );
}
