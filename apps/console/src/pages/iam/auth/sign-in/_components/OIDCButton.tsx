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
import { Button, Google, Microsoft } from "@probo/ui";
import type { ComponentProps } from "react";
import { useFragment } from "react-relay";
import { useSearchParams } from "react-router";
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
  const [searchParams] = useSearchParams();
  const safeContinueUrl = useSafeContinueUrl();
  const provider = useFragment(fragment, providerRef);
  const Icon = providerIcons[provider.name];
  const organizationId = searchParams.get("organization-id");

  return (
    <Button
      variant="secondary"
      className="w-full h-10"
      onClick={() => {
        const loginURL = new URL(provider.loginURL, window.location.origin);
        loginURL.searchParams.set("continue", safeContinueUrl.toString());
        if (organizationId) {
          loginURL.searchParams.set("organization_id", organizationId);
        }

        window.location.href = loginURL.toString();
      }}
    >
      <span className="flex items-center gap-2">
        {Icon && <Icon width={18} height={18} />}
        {__(`Sign in with ${provider.name.charAt(0).toUpperCase() + provider.name.slice(1)}`)}
      </span>
    </Button>
  );
}
