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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { GoogleLogo } from "@probo/ui/src/v2/GoogleLogo/GoogleLogo";
import { MicrosoftLogo } from "@probo/ui/src/v2/MicrosoftLogo/MicrosoftLogo";
import type { ComponentProps } from "react";
import { useTranslation } from "react-i18next";
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
  google: GoogleLogo,
  microsoft: MicrosoftLogo,
};

export function OIDCButton({
  providerRef,
  continueURL,
}: {
  providerRef: OIDCButtonFragment$key;
  continueURL?: string;
}) {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const safeContinueUrl = useSafeContinueUrl();
  const provider = useFragment(fragment, providerRef);
  const Icon = providerIcons[provider.name];
  const organizationId = searchParams.get("organization-id");
  const targetContinue = continueURL ?? safeContinueUrl.toString();
  const providerLabel
    = provider.name.charAt(0).toUpperCase() + provider.name.slice(1);

  return (
    <Button
      variant="soft"
      color="neutral"
      size={3}
      className="min-w-0 flex-1"
      aria-label={t("oidcButton.signInWith", { provider: providerLabel })}
      iconStart={Icon ? <Icon className="size-5" /> : undefined}
      onClick={() => {
        const loginURL = new URL(provider.loginURL, window.location.origin);
        loginURL.searchParams.set("continue", targetContinue);
        if (organizationId) {
          loginURL.searchParams.set("organization_id", organizationId);
        }

        window.location.href = loginURL.toString();
      }}
    >
      {providerLabel}
    </Button>
  );
}
